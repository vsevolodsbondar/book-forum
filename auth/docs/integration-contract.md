# Integration Contract / Team Handoff

This document is the handoff boundary between the AUTH work and the BACKEND / FRONTEND work.

The goal is that each team can implement against a stable contract without depending on AUTH internals.

---

# What AUTH delivers

AUTH will provide BACKEND with a private HTTP service that supports:

```text
POST /v1/register
POST /v1/login
POST /v1/session/validate
POST /v1/logout
```

AUTH will also provide:

- documented JSON request/response shapes;
- stable HTTP status codes and error codes;
- Argon2id password storage;
- user uniqueness enforcement;
- UUIDv4 session identifiers;
- opaque random session credentials;
- session expiration and revocation;
- SQLite migrations for AUTH-owned data;
- automated tests for core authentication behavior.

AUTH owns its database. Neither FRONTEND nor BACKEND should read or mutate AUTH tables directly.

---

# Contract with BACKEND

## What AUTH expects from BACKEND

### 1. AUTH stays private

BACKEND is expected to be the only application service calling AUTH.

The AUTH container should not publish a browser-facing host port. It is intended to remain reachable only by BACKEND on the private application network.

### 2. BACKEND exposes the public auth routes

Suggested public routes:

```text
POST /api/auth/register
POST /api/auth/login
POST /api/auth/logout
GET  /api/auth/me
```

These names are BACKEND's public contract. They do not need to match AUTH's internal paths exactly.

### 3. BACKEND forwards minimal typed payloads

BACKEND should parse public requests and construct a new request to AUTH.

For registration, forward only:

```json
{
	"email": "...",
	"username": "...",
	"password": "..."
}
```

For login, forward only:

```json
{
	"email": "...",
	"password": "..."
}
```

Do not blindly proxy arbitrary browser headers, query parameters, or cookies to AUTH.

### 4. BACKEND owns the browser-facing session cookie

After a successful AUTH login, BACKEND receives an opaque token and expiry.

Sessions expire after 7 days of inactivity or 30 days after creation by
default. AUTH enforces both limits. The returned `expires_at` represents
absolute expiry; BACKEND uses it for cookie expiry rather than hardcoding
the duration. Activity does not extend absolute expiry. AUTH can reject an
inactive session before its cookie expires.

BACKEND must set the token as a cookie rather than returning it to frontend JavaScript.

Intended secure deployment cookie configuration:

```text
Name:     __Host-session
HttpOnly: true
Secure:   true
SameSite: Lax
Path:     /
Domain:   not set
Max-Age / Expires: match AUTH session expiry
```

For local development over plain HTTP, the `Secure` cookie attribute may need to be disabled and the `__Host-` prefix may therefore not be usable.

The intended secure deployment configuration uses the hardened settings above. Development-only cookie settings must not be treated as the secure deployment configuration.

### 5. BACKEND validates sessions through AUTH

For protected requests:

```text
browser cookie
   -> BACKEND
   -> opaque token
   -> AUTH /v1/session/validate
   -> user_id / session identity
```

BACKEND must not attempt to recreate AUTH's session lookup by querying `auth.sqlite`.

### 6. BACKEND owns application authorization

AUTH answers:

```text
Who is this user?
Is this session valid?
```

BACKEND answers:

```text
May this user edit this post?
May this user access this resource?
What application data belongs to this user?
```

For logout, BACKEND calls AUTH and clears the browser cookie regardless of
the result. AUTH returns indistinguishable `204 No Content` responses for
missing, malformed, unknown, expired, and active tokens; only an internal
AUTH failure is surfaced as an error. Clearing the cookie after a failure
does not guarantee that the server-side session was revoked.

### 7. BACKEND stores `user_id` as the external identity reference

Application/profile tables may reference AUTH's stable numeric user ID:

```text
profile.user_id = 42
post.user_id = 42
comment.user_id = 42
```

Do not copy password hashes, session hashes, or AUTH's credential data into the BACKEND DB.

### 8. BACKEND must not log secrets

Do not log complete bodies for:

```text
/api/auth/register
/api/auth/login
```

Do not log:

- plaintext passwords;
- session cookie values;
- AUTH login response session tokens;
- Authorization headers used for internal session validation.

### 9. BACKEND handles public-edge protections

BACKEND is expected to own:

- public request body limits;
- public auth endpoint rate limiting / throttling;
- HTTPS termination or deployment behind HTTPS;
- CSRF protection / origin checks for cookie-authenticated state-changing routes;
- public error mapping.

AUTH independently limits registration and login bodies to 8 KiB (8192 bytes)
and concurrent Argon2id operations to 2 by default. These endpoints can return
`413 request_too_large` or `503 service_busy` using the documented error
envelope. Hashing saturation does not consume capacity for session validation
or logout. See [AUTH resource limits](security.md#resource-limits) for details.

---

# Contract with FRONTEND

## What FRONTEND receives

FRONTEND talks to BACKEND only.

Suggested registration request:

```http
POST /api/auth/register
Content-Type: application/json
```

```json
{
	"email": "alice@example.com",
	"username": "Alice",
	"password": "..."
}
```

Suggested login request:

```http
POST /api/auth/login
Content-Type: application/json
```

```json
{
	"email": "alice@example.com",
	"password": "..."
}
```

FRONTEND should receive safe user/account data, not password/session internals.

Example successful login public response:

```json
{
	"user": {
		"id": 42,
		"email": "alice@example.com",
		"username": "Alice"
	}
}
```

The session token is delivered by BACKEND as an HttpOnly cookie and is intentionally unavailable to frontend JavaScript.

## What AUTH/BACKEND expect from FRONTEND

### Registration form

Required fields:

```text
email
username
password
```

Email requirements:

- One plain ASCII email address, without a display-name wrapper.
- Maximum length after trimming: 254 bytes, with at most 64 bytes before `@`.
- AUTH trims surrounding whitespace and lowercases the address.
- Email identities are case-insensitively unique.
- FRONTEND may validate these rules for UX; AUTH remains authoritative.

See [email requirements](api.md#email-requirements) for the full validation rules.

Username requirements

- Length after trimming: 3 to 30 characters.
- Allowed characters: ASCII letters (`A–Z`, `a–z`), digits (`0–9`), and underscores (`_`).
- Surrounding whitespace is trimmed before validation and storage.
- Display casing is preserved.
- Usernames are unique regardless of ASCII letter casing: `Alice` and `alice` conflict.

See [username requirements](api.md#username-requirements) for the full validation rules.

Password requirements:

- 15 to 128 Unicode code points.
- Spaces and Unicode characters are allowed.
- Passwords are never truncated.
- No composition rules are required.
- FRONTEND may validate these requirements for UX, but AUTH remains authoritative.

See [password requirements](api.md#password-requirements) for the full validation rules.

### Login form

Required fields:

```text
email
password
```

Username is not a login credential.

### Client-side validation

FRONTEND may validate fields for UX, but server validation remains authoritative.

### Cookies

FRONTEND must not attempt to:

- read the session token;
- store the session token in `localStorage`;
- store the session token in `sessionStorage`;
- manually construct authentication headers from the cookie.

The browser should send the HttpOnly cookie automatically on same-origin requests.

### Logout

FRONTEND calls:

```text
POST /api/auth/logout
```

BACKEND handles revocation and cookie deletion.

---

# Handoff checklist

## AUTH -> BACKEND

- [x] Internal AUTH base URL / container service name provided.
- [ ] `/v1/register` implemented and documented.
- [ ] `/v1/login` implemented and documented.
- [ ] `/v1/session/validate` implemented and documented.
- [ ] `/v1/logout` implemented and documented.
- [ ] Error codes/status codes finalized.
- [ ] Session expiry duration finalized/configurable.
- [ ] Example requests available.

## BACKEND -> AUTH

- [ ] Public auth routes agreed.
- [ ] Cookie name/config agreed.
- [ ] BACKEND does not expose raw AUTH token to frontend JS.
- [ ] BACKEND does not query AUTH DB.
- [ ] Auth route request logging disabled/redacted.
- [ ] Public rate limiting strategy agreed.

## FRONTEND -> BACKEND

- [ ] Registration form fields agreed.
- [ ] Login form fields agreed.
- [ ] Public error shape agreed.
- [ ] Logout behavior agreed.
- [ ] Frontend does not require direct access to session token.
