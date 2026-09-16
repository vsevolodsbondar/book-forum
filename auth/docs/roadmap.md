# AUTH v0 Roadmap

## Goal

Deliver the smallest complete authentication service that satisfies the assignment and provides a clean foundation for a future standalone authentication service.

v0 is complete when a user can:

```text
register
login
remain authenticated via an expiring cookie-backed server session
access protected backend functionality
logout
```

AUTH itself is private and is integrated through BACKEND.

---

## Phase 0 - Foundation

### Deliverables

- [x] Create AUTH package/service structure.
- [x] Configure SQLite connection with foreign keys enabled.
- [ ] Add migration mechanism.
- [x] Add `0001_init.sql`.
- [x] Add `users` table.
- [x] Add `sessions` table.
- [ ] Add schema tests / migration startup test.
- [ ] Verify the [schema guarantees](schema.md): automatically assigned user IDs are not reused, email and ASCII case-insensitive username uniqueness are enforced, user deletion cascades to sessions, and required indexes exist.
- [ ] Add configuration for absolute session lifetime (default: 30 days), idle timeout (default: 7 days), and Argon2id parameters.
- [ ] Reject non-positive session timeout configuration; zero cannot disable either timeout.
- [ ] Add `ARGON2_MAX_CONCURRENCY` configuration (default: 2) and reject non-positive values.

### Done when

A fresh AUTH instance can create or open its database and successfully apply the initial schema.

---

## Phase 1 - Password subsystem

### Deliverables

- [ ] Implement Argon2id hashing through `x/crypto/argon2`.
- [ ] Generate a fresh random salt per password.
- [ ] Encode a self-describing password hash string.
- [ ] Parse encoded hashes.
- [ ] Validate parsed cost parameters against accepted bounds.
- [ ] Verify passwords using constant-time comparison.
- [ ] Share one non-blocking concurrency limit across registration hashing, login verification, and unknown-email dummy verification.
- [ ] Return `503 service_busy` when no hashing slot is available, without queuing or starting Argon2id; release slots after completion, including failures.
- [ ] Unit-test correct, incorrect, malformed, and differently salted passwords.

### Done when

The password package can safely round-trip:

```text
password -> encoded Argon2id hash -> verify true/false
```

without the HTTP or database layer needing to know Argon2id implementation details.

---

## Phase 2 - Registration

### Internal endpoint

```text
POST /v1/register
```

### Deliverables

- [ ] Decode JSON safely.
- [ ] Enforce an 8192-byte body limit while reading, regardless of `Content-Length`, returning `413 request_too_large` when exceeded.
- [ ] Validate with `go-playground/validator`.
- [ ] Enforce the [password requirements](api.md#password-requirements), including length of 15 to 128 Unicode code points.
- [ ] Preserve passwords exactly as submitted without trimming or normalization.
- [ ] Validate and normalize email according to the [email requirements](api.md#email-requirements).
- [ ] Validate and trim username according to the [username requirements](api.md#username-requirements), preserving display casing.
- [ ] Hash the password with Argon2id.
- [ ] Insert the user using parameterized SQL.
- [ ] Handle email and username uniqueness conflicts.
- [ ] Return only safe user fields.
- [ ] Add integration tests.

### Done when

Unique users can register and duplicate identities are rejected correctly.

Registration does not automatically log the user in.

---

## Phase 3 - Login and session creation

### Internal endpoint

```text
POST /v1/login
```

### Deliverables

- [ ] Look up the user by normalized email.
- [ ] Enforce the same 8192-byte body limit and `413 request_too_large` response as registration.
- [ ] Verify the Argon2id password hash.
- [ ] Return generic invalid-credentials errors.
- [ ] Prepare one reusable dummy Argon2id hash using current hashing parameters and verify against it for unknown emails before returning `invalid_credentials`.
- [ ] Generate a UUIDv4 session ID.
- [ ] Generate a separate 32-byte random session secret.
- [ ] Compute SHA-256 over the raw session secret.
- [ ] Store only the resulting 32-byte hash as a BLOB.
- [ ] Base64URL-encode the raw session secret to produce the opaque session token.
- [ ] Initialize `last_seen_at` at creation and store fixed absolute `expires_at` using the configured lifetime (default: 30 days).
- [ ] Return the encoded session token and expiry to BACKEND only.
- [ ] Add tests proving that neither the raw session secret nor the encoded session token is persisted.

### Done when

Valid credentials create a server-side session and invalid credentials do not.

---

## Phase 4 - Session validation

### Internal endpoint

```text
POST /v1/session/validate
```

### Deliverables

- [ ] Read the opaque session token from the internal `Authorization` header.
- [ ] Base64URL-decode the token.
- [ ] Reject malformed tokens or tokens whose decoded value is not exactly 32 bytes.
- [ ] Compute SHA-256 over the decoded raw session secret.
- [ ] Look up the indexed `token_hash`.
- [ ] Reject sessions reaching absolute expiry or the configured idle timeout (default: 7 days), using server-side time.
- [ ] Update `last_seen_at` on every successful validation without extending `expires_at`.
- [ ] Return minimal identity information: user ID, username, session ID, and expiry.
- [ ] Add tests for valid, random, malformed, and expired tokens.
- [ ] Test acceptance just before and rejection exactly at both timeout boundaries.
- [ ] Test successful activity updates and unchanged absolute expiry.
- [ ] Verify that expired sessions are rejected before activity updates and cannot be revived.

### Done when

BACKEND can reliably convert a browser session credential into an authenticated `user_id` without accessing AUTH storage directly.

---

## Phase 5 - Logout

### Internal endpoint

```text
POST /v1/logout
```

### Deliverables

- [ ] Read and Base64URL-decode the presented session token.
- [ ] Compute SHA-256 over the decoded raw session secret.
- [ ] Delete the matching session.
- [ ] Make repeated logout safe and idempotent at the public layer.
- [ ] Verify that a deleted session token can no longer validate.
- [ ] Return `204 No Content` without revealing token state for missing, malformed, unknown, expired, and active tokens.
- [ ] Delete matching active or expired session rows when possible; return `500 internal_error` only for database or unexpected internal failures.
- [ ] Test each logout token outcome, repeated logout, expired-row cleanup, and database-failure handling.

### Done when

Logout immediately revokes the server-side session.

---

## Phase 6 - BACKEND integration

AUTH handoff requirements are documented in `integration-contract.md`.

### Integration acceptance

- [ ] BACKEND forwards registration and login payloads to AUTH.
- [ ] BACKEND sets the browser session cookie.
- [ ] BACKEND validates protected requests through AUTH.
- [ ] BACKEND clears the browser cookie on logout.
- [ ] BACKEND uses the returned AUTH `user_id` for application ownership.
- [ ] BACKEND never reads the AUTH database directly.
- [ ] Sensitive authentication bodies, headers, and tokens are not logged.

---

## Phase 7 - FRONTEND integration

### Integration acceptance

- [ ] Registration form sends email, username, and password to BACKEND.
- [ ] Registration form communicates the 15 to 128 Unicode code point password requirement.
- [ ] Login form sends email and password to BACKEND.
- [ ] FRONTEND does not trim or normalize passwords before submission.
- [ ] FRONTEND never receives or stores the session token in JavaScript.
- [ ] Logout calls BACKEND.
- [ ] Protected requests rely on the browser-managed session cookie.

---

## Phase 8 - Final security and test pass

- [ ] Run unit tests.
- [ ] Run integration tests.
- [ ] Test concurrent duplicate registration behavior.
- [ ] Test absolute and idle session expiration end-to-end, including inactivity expiry before cookie expiry.
- [ ] Confirm that plaintext passwords, password hashes, session credentials, sensitive request/response bodies, session cookie values, and internal Authorization headers are excluded or redacted from logs.
- [ ] Verify all four endpoints match the [HTTP contract](api.md): response fields, status and error codes, JSON error envelope, and empty `204` bodies.
- [ ] Test that absolute-lifetime configuration changes affect newly created sessions only, while current idle-timeout configuration applies to existing sessions.
- [ ] Benchmark Argon2id in the AUTH container.
- [ ] Benchmark the default Argon2id concurrency of 2 against container memory and latency constraints.
- [ ] Test bodies at and above the 8192-byte limit on registration and login, including requests without `Content-Length`.
- [ ] Test shared Argon2id capacity across endpoints, including unknown-email dummy verification, immediate `503 service_busy` on saturation, and slot release after failure.
- [ ] Verify session validation and logout remain available while hashing slots are occupied.
- [ ] Confirm indexes are used for session lookup.
- [ ] Confirm SQLite foreign keys are enabled on actual database connections.
- [ ] Document the final Argon2id and session configuration.

---

# v0 Definition of Done

```text
REGISTER
email + username + password
        -> AUTH
        -> validated + normalized where applicable
        -> Argon2id
        -> users row

LOGIN
email + password
        -> AUTH verifies Argon2id
        -> UUIDv4 session ID
        -> 256-bit random session secret
        -> SHA-256(secret) stored
        -> Base64URL token returned only to BACKEND
        -> BACKEND sets HttpOnly expiring cookie

PROTECTED REQUEST
cookie
        -> BACKEND
        -> AUTH validates token
        -> user_id returned
        -> BACKEND performs application authorization

LOGOUT
cookie token
        -> BACKEND
        -> AUTH revokes session
        -> BACKEND clears cookie
```

# Post-v0 backlog

Do not delay v0 for these:

- email verification;
- password reset;
- password change with session revocation;
- login throttling inside AUTH as defense in depth;
- session/device management;
- revoke all sessions;
- MFA / TOTP;
- WebAuthn/passkeys;
- recovery codes;
- audit/security events;
- OAuth/OIDC;
- Argon2id parameter auto-upgrade after successful login.
- evaluate NFC password normalization using `golang.org/x/text/unicode/norm`, including compatibility with existing exact-input password hashes;
- common and compromised password checking using [Pwned Passwords](https://haveibeenpwned.com/Passwords), choosing offline data or API integration.
