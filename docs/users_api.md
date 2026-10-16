# User API

Covers the User endpoints: register, login, logout, list, get, update, delete — and the shared error response format.

## Errors

Every error response uses the same JSON envelope, produced by the shared `GlobalErrorHandler`:

```json
{
  "status": 404,
  "message": "user not found"
}
```

The HTTP status code matches `status` in the body. Sentinel errors map to status codes like this:

| Sentinel error | HTTP status | Meaning |
|---|---|---|
| `ErrInvalidInput`, `ErrBadRequest` | 400 | Malformed request (bad JSON, unknown fields, body too large, invalid `id`/`page`/`size`) |
| `ErrInvalidID`, `ErrEmptyUsername`, `ErrUsernameTooLong`, `ErrDescTooLong` | 400 | Field validation failure |
| `ErrExpiredSession` | 401 | Session cookie missing/invalid/expired |
| `ErrWrongCredentials` | 401 | Login failed (wrong email or password) |
| `ErrForbidden` | 403 | Authenticated, but not the owner of this account |
| `ErrUserNotFound` | 404 | The referenced user doesn't exist |
| `ErrConflict` | 409 | The auth service reports a conflict (e.g. email already registered) |
| anything else (`ErrGetUser`, `ErrGetUsers`, `ErrCreateUser`, `ErrUpdateUser`, `ErrDeleteUser`, `ErrGetRowsAffected`, `ErrAuthService`) | 500 | Unexpected internal failure; message not leaked to the client |

## Authentication

Credentials and sessions are owned by the separate auth service. This service never stores email or passwords.

- `POST /login` creates a session in the auth service and sets the **`session_id`** cookie (`HttpOnly`, `SameSite=Lax`, `Path=/`, expiry taken from the auth service). The session token is only ever sent in this cookie, never in a JSON body.
- Endpoints marked **session required** read the cookie and validate it against the auth service.
- A user can only update or delete **their own** account (`ErrForbidden` otherwise).

| Endpoint | Session required |
|---|---|
| `POST /login`, `POST /users`, `GET /users`, `GET /users/{id}` | no |
| `POST /logout`, `PATCH /users/{id}`, `DELETE /users/{id}` | yes |

## Login

`POST /login`

**Request body**

```json
{
  "email": "user@example.com",
  "password": "secret123"
}
```

Unknown fields are rejected (`ErrInvalidInput`, 400).

**Behavior**

- Forwards the credentials to the auth service, which verifies them and creates the session.
- Rejects with `ErrWrongCredentials` (401) if the email or password is wrong. The auth service gives the same answer for an unknown email and a wrong password.
- If the auth service is unavailable or overloaded, the request is retried a few times and then fails with `500`.

**Response**

`200 OK`. Sets the `session_id` cookie. The body is the session id as a JSON string.

## Logout

`POST /logout`

Requires a valid session.

**Behavior**

- Rejects with `ErrExpiredSession` (401) if the cookie is missing.
- Asks the auth service to delete the session, then clears the cookie in the browser.
- If the auth service call fails, the error is returned and the cookie is **not** cleared.

**Response**

`204 No Content`.

## Create a user

`POST /users`

No session required — this is how an account is created in the first place.

**Request body**

```json
{
  "email": "user@example.com",
  "password": "secret123",
  "username": "user123",
  "profilepic": "",
  "name": "user",
  "description": ""
}
```

Unknown fields are rejected (`ErrInvalidInput`, 400). The body is limited to 1 MB (`ErrBadRequest`, 400).

**Behavior**

- Registers the account in the auth service first. `email` and `password` go there and are never stored locally.
- Then creates the local profile (`username`, `profilepic`, `name`, `description`) using the `id` returned by the auth service, so both services agree on the user's identity.
- If the auth service rejects the registration, its error is returned (`ErrInvalidInput` for 400, `ErrConflict` for 409, otherwise `500`).
- If the local insert fails (for example the username is already taken), the result is `500`.

> **Known gaps:** `validateSubmission` (username length, description length) exists in the service but is not called, so registration input is not validated. If the local insert fails after the auth registration succeeded, the auth account remains without a local profile.

**Response**

`201 Created`. Returns the created user's public profile (`id`, `username`, `profilepic`, `name`, `description`, `created_at`, `last_seen`). Never includes `email` or credentials.

## List users

`GET /users`

No session required.

**Pagination**

- `page` — defaults to `1`. Must be `>= 1`.
- `size` — defaults to `20`. Must be `> 0`.

Example: `GET /users?page=2&size=10`. Invalid values are rejected with `ErrInvalidInput` (400). Users are ordered by `id`.

**Response**

`200 OK`.

```json
{
  "users": [
    { "id": 1, "username": "egor123", "profilepic": "" }
  ],
  "page": 1,
  "size": 20,
  "total": 1
}
```

`total` is the number of users in the whole table, not on the page. A page past the end returns `"users": []`. A missing profile picture is returned as an empty string.

## Get a user

`GET /users/{id}`

No session required — profiles are public.

**Behavior**

- Rejects with `ErrInvalidInput` (400) if `id` isn't a positive integer.
- Rejects with `ErrUserNotFound` (404) if no user exists with this `id`.

**Response**

`200 OK`. Returns the user's public profile (`id`, `username`, `profilepic`, `name`, `description`, `created_at`, `last_seen`). Missing optional fields are returned as empty strings.

## Update a user

`PATCH /users/{id}`

Requires a valid session.

**Request body** — all fields optional, only the fields present are changed:

```json
{
  "username": "newname",
  "profilepic": "",
  "name": "New Name",
  "description": "Updated bio"
}
```

Only profile fields can be changed. `email` and `password` are managed by the auth service. Unknown fields are rejected (400); the body is limited to 1 MB.

**Behavior**

- Rejects with `ErrExpiredSession` (401) if the cookie is missing or the session is invalid.
- Rejects with `ErrForbidden` (403) if the session's user doesn't match `id` in the URL.
- If provided, `username` must be non-empty and at most 32 characters, and `description` at most 500 (`400` otherwise).
- Rejects with `ErrUserNotFound` (404) if `id` doesn't refer to an existing user.
- Changing `username` to one that is already taken fails with `500`.

**Response**

`200 OK`, empty body.

## Delete a user

`DELETE /users/{id}`

Requires a valid session.

**Behavior**

- Rejects with `ErrExpiredSession` (401) if the cookie is missing or the session is invalid.
- Rejects with `ErrForbidden` (403) if the session's user doesn't match `id` in the URL.
- Rejects with `ErrUserNotFound` (404) if `id` doesn't refer to an existing user.
- Posts and comments by the deleted user are kept: `author_id`/`user_id` on them is set to `null`. The user's reactions are deleted (`ON DELETE CASCADE`).

> **Known gap:** only the local profile is deleted. The account and session in the auth service are not touched.

**Response**

`204 No Content`.
