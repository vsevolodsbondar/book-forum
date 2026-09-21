# User API

Covers the User endpoints: create (register), get, update, delete.

## Errors

> **Note:** unlike the Comment API, this service does not yet have a unified `GlobalErrorHandler`. Errors are currently written with `http.Error(w, err.Error(), <status>)`, so the response body is a plain text string, not a JSON envelope. The underlying Go error (wrapped with `%w`) is what determines the status code chosen by the handler — see the table below.

| Condition | HTTP status | Meaning |
|---|---|---|
| Malformed JSON body | 400 | Request body isn't valid JSON, or contains unknown fields |
| Request body exceeds `maxReqBodySize` (1 MB) | 413 | Body too large |
| `id` path value isn't a positive integer | 400 | `ErrInvalidID` |
| Session cookie missing/invalid | 401 | `helper.ExtractSessionCookie` or `auth.ValidateSession` failed |
| Session doesn't belong to the target `id` | 403 | Caller isn't the account owner |
| User not found (`GetUser`) | 404 | No row with this `id` |
| Auth service rejects registration | 400 | e.g. email/username already taken |
| Any other repository/service error | 400 or 500 (see per-endpoint notes below) | Currently returned via `err.Error()` — message may leak internal detail; not yet sanitized |

## Create a user

`POST /api/users`

No session required — this is how an account is created in the first place.

**Request body** (`model.UserDTO`)

```json
{
  "email": "egor@example.com",
  "password": "secret123",
  "username": "egor123",
  "profilepic": "",
  "name": "Egor",
  "description": ""
}
```

Unknown fields are rejected (`DisallowUnknownFields`).

**Behavior**

- Calls the auth service (`Auth.RegisterUser`) to create the credentialed account and obtain the new user's `id`. `email`/`password` are forwarded there and never stored in this service's own database.
- The local profile row is created with the `id` returned by the auth service, so both services agree on the user's identity.
- If the auth service call fails, `CreateUser` returns early with a wrapped error → `400 Bad Request`.
- If the local insert fails, same → `400 Bad Request` (this includes cases that are arguably server errors, e.g. a DB connectivity issue — see caveat below).

> **Caveat:** `validateSubmission` (username length/emptiness, description length) is defined in the service but is **not currently called** by `CreateUser`. Input validation for creation isn't enforced yet — worth confirming this is intentional before relying on it.

**Response**

`201 Created`. Returns the created user's public profile (`id`, `username`, `profilepic`, `name`, `description`, `created_at`, `last_seen`). Never includes `email` or credential data — those live only in the auth service.

## Get a user

`GET /api/users/{id}`

No session required — profiles are public.

**Behavior**

- `id` must be a positive integer, or `400 Bad Request`.
- Returns `404 Not Found` if no user exists with this `id`.

**Response**

`200 OK`. Returns the user's public profile (`id`, `username`, `profilepic`, `name`, `description`, `created_at`, `last_seen`).

## Update a user

`PUT /api/users/{id}`

Requires a valid session.

**Request body** (`model.UserUpdateInfo` — all fields optional/pointer-based; only fields present are changed, via `COALESCE` in the update query)

```json
{
  "username": "newname",
  "profilepic": "",
  "name": "New Name",
  "description": "Updated bio"
}
```

Only profile fields can be changed here — `email`/`password` are managed by the auth service, not this endpoint.

**Behavior**

- `id` must be a positive integer, or `400 Bad Request`.
- Requires a session cookie (`helper.ExtractSessionCookie`) and a valid session (`auth.ValidateSession`) — missing/invalid → `401 Unauthorized`.
- The session's user must match the `id` in the URL — otherwise `403 Forbidden`. A user can only update their own profile.
- If provided, `username` must be non-empty and ≤ 32 characters; `description` must be ≤ 500 characters — violations return a validation error (currently surfaced as `500 Internal Server Error`, see caveat).
- Returns an error if `id` doesn't refer to an existing user (currently surfaced as `500 Internal Server Error`, see caveat).

> **Caveat:** `UpdateUser` in the handler currently maps *all* service errors to `500 Internal Server Error`, including validation failures and "not found" — these are arguably `400`/`404` respectively. Worth aligning with the status-code table once a unified error writer is in place.

**Response**

`200 OK` on success, empty body.

## Delete a user

`DELETE /api/users/{id}`

Requires a valid session.

**Behavior**

- `id` must be a positive integer, or `400 Bad Request`.
- Requires a session cookie and a valid session — missing/invalid → `401 Unauthorized`.
- The session's user must match the `id` in the URL — otherwise `403 Forbidden`. A user can only delete their own account.
- Returns an error if `id` doesn't refer to an existing user (currently surfaced as `500 Internal Server Error`, see caveat above).
- Posts and comments authored by the deleted user are not removed — `author_id`/`user_id` on those rows is set to `null` (`ON DELETE SET NULL` in the schema), so existing content stays visible with an anonymized author.

**Response**

`204 No Content` on success.
