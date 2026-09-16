# AUTH Internal API

> This API is **internal**. It is intended for BACKEND, not browsers.

By default, AUTH is reachable inside the container network at:

```text
http://auth:8081/v1
```

The listening port is configured through AUTH's `PORT` environment variable and defaults to `8081`. It is an internal service detail and is not part of the browser-facing contract.

## Conventions

- JSON request and response bodies are used unless the endpoint returns `204 No Content`.
- JSON requests use `Content-Type: application/json`.
- Unix timestamps are expressed as seconds since the Unix epoch.
- `now` means the current server-side Unix timestamp in seconds.
- Session tokens use unpadded Base64URL encoding (`base64.RawURLEncoding` in Go).
- Authentication failures do not reveal whether an email exists.
- Raw passwords and raw session tokens must never be logged.
- Error bodies use stable machine-readable codes.
- Registration and login request bodies are limited to 8 KiB (8192 bytes),
  enforced while reading regardless of `Content-Length`.
- Registration and login share a non-blocking Argon2id concurrency limit
  (default: 2 operations). See [resource limits](security.md#resource-limits).

Error bodies use the following shape:

```json
{
	"error": {
		"code": "invalid_credentials",
		"message": "invalid email or password"
	}
}
```

---

## `POST /v1/register`

Creates a user account.

### Request

```json
{
	"email": "Alice@Example.com",
	"username": "Alice",
	"password": "correct horse battery staple"
}
```

### Email requirements

- Accept one plain ASCII email address, without a display-name wrapper.
- Quoted local-parts and IP-address domain literals are not supported.
- Trim surrounding whitespace before validation.
- Maximum length after trimming: 254 bytes.
- Maximum local-part length (before `@`): 64 bytes.
- Validate email syntax.
- Lowercase the entire address before storage or login lookup.
- Email identities are case-insensitively unique.
- Internationalized email addresses are outside v0 scope.
- Syntax validation does not verify mailbox ownership.

### Username requirements

- Length after trimming: 3 to 30 characters.
- Allowed characters: ASCII letters (`A–Z`, `a–z`), digits (`0–9`), and underscores (`_`).
- Surrounding whitespace is trimmed before validation and storage.
- Display casing is preserved.
- Usernames are unique regardless of ASCII letter casing: `Alice` and `alice` conflict.

### Password requirements

- Minimum length: 15 Unicode code points.
- Maximum length: 128 Unicode code points.
- Spaces and Unicode characters are allowed.
- Passwords are never truncated.
- No uppercase, lowercase, number, or symbol composition rules are required.
- Passwords are processed exactly as submitted and are not trimmed or normalized.

### AUTH responsibilities

1. Validate all fields.
2. Normalize the email using trim + lowercase.
3. Trim the username and apply username rules.
4. Hash the password with Argon2id using a fresh random salt.
5. Insert the user and allow SQLite uniqueness constraints to remain the final authority.
6. Never return `password_hash`.

### Success - `201 Created`

```json
{
	"user": {
		"id": 42,
		"email": "alice@example.com",
		"username": "Alice",
		"created_at": 1789560000
	}
}
```

### Errors

- **`400 invalid_request`** - malformed or invalid registration data.
- **`409 identity_conflict`** - email or username is already in use.
- **`413 request_too_large`** - request body exceeds 8192 bytes.
- **`500 internal_error`** - unexpected server failure.
- **`503 service_busy`** - no Argon2id slot is available; no hashing operation is queued or started.

Registration does **not** create a session in v0.

---

## `POST /v1/login`

Verifies credentials and creates a login session.

### Request

```json
{
	"email": "alice@example.com",
	"password": "correct horse battery staple"
}
```

### AUTH responsibilities

1. Normalize the email.
2. Load the matching user.
3. Verify the Argon2id password hash.
4. Generate a UUIDv4 session ID.
5. Generate a separate 32-byte session secret using `crypto/rand`.
6. Compute `SHA-256(secret)` and store only the resulting hash in `sessions.token_hash`.
7. Base64URL-encode the raw session secret to produce the opaque session token.
8. Initialize `last_seen_at` to the creation time and set `expires_at` to that
   time plus the configured absolute lifetime (default: 30 days).
9. Return the opaque session token only to BACKEND over the internal service boundary.

### Success - `200 OK`

```json
{
	"user": {
		"id": 42,
		"email": "alice@example.com",
		"username": "Alice"
	},
	"session": {
		"id": "6a79f46f-e3a1-4f22-8b89-13ec5a9dbc31",
		"token": "base64url-encoded-opaque-secret",
		"expires_at": 1790164800
	}
}
```

`session.token` is sensitive. BACKEND uses it to create the browser-facing session cookie and must not return it in frontend JSON.

`session.expires_at` is the fixed absolute expiry. Sessions also expire after
the configured idle timeout (default: 7 days). Activity never extends the
absolute expiry. See [session lifetime](security.md#session-lifetime) for configuration.

### Errors

- **`400 invalid_request`** - malformed JSON or invalid request data.
- **`401 invalid_credentials`** - email or password is incorrect.
- **`413 request_too_large`** - request body exceeds 8192 bytes.
- **`500 internal_error`** - unexpected server failure.
- **`503 service_busy`** - no Argon2id slot is available, including for unknown-email dummy verification; no verification operation is queued or started.

AUTH returns the same `invalid_credentials` response whether the email is unknown or the password is incorrect.

---

## `POST /v1/session/validate`

Validates an opaque session credential.

### Request

```http
Authorization: Bearer <opaque-session-token>
```

No JSON body is required.

### AUTH responsibilities

1. Extract the bearer token.
2. Base64URL-decode the token.
3. Reject malformed tokens or tokens with an unexpected decoded length.
4. Compute SHA-256 of the decoded raw session-secret bytes.
5. Look up the resulting hash in `sessions.token_hash`.
6. Require `expires_at > now` and `last_seen_at + idle_timeout > now`, where
   `idle_timeout` is the configured idle timeout in seconds (default: 7 days).
7. Update `last_seen_at` to record successful validation only after both
   timeout checks pass. Never extend `expires_at` or revive an expired session.
8. Return the identity associated with the session.

### Success - `200 OK`

```json
{
	"session": {
		"id": "6a79f46f-e3a1-4f22-8b89-13ec5a9dbc31",
		"expires_at": 1790164800
	},
	"user": {
		"id": 42,
		"username": "Alice"
	}
}
```

### Errors

- **`401 invalid_session`** - session token is missing, malformed, expired, or not recognized.
- **`500 internal_error`** - unexpected server failure.

Example `401 Unauthorized` response:

```json
{
	"error": {
		"code": "invalid_session",
		"message": "invalid or expired session"
	}
}
```

---

## `POST /v1/logout`

Revokes the presented session.

### Logout behavior

Logout is idempotent. AUTH returns `204 No Content` for all client-token
outcomes:

- no `Authorization` header;
- malformed or incorrectly encoded token;
- valid token whose session does not exist;
- expired session, including when its row still exists;
- active session successfully deleted.

The `204` response body is always empty. BACKEND should clear the browser cookie
regardless of which outcome occurred.

AUTH returns `500 internal_error` only when it cannot safely complete the
request because of a database or unexpected internal failure. AUTH must not
reveal token validity, session existence, or expiration status in the client
response.

### Request

```http
Authorization: Bearer <opaque-session-token>
```

No JSON body is required.

### AUTH responsibilities

1. Extract and Base64URL-decode the presented token.
2. Hash the decoded raw session-secret bytes exactly as session validation does.
3. Delete the matching session row.
4. Return `204 No Content`.

### Success - `204 No Content`

No response body.

BACKEND should clear the browser cookie regardless of whether the previous session token was already expired or missing.

### Errors

- **`500 internal_error`** - unexpected server failure.

---

## Possible post-v0 endpoints

These endpoints are intentionally not part of the initial contract:

```text
POST   /v1/password/change
POST   /v1/password/reset/request
POST   /v1/password/reset/confirm
POST   /v1/email/verify
GET    /v1/sessions
DELETE /v1/sessions/{session_id}
DELETE /v1/sessions
```
