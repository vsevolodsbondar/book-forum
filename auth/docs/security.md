# Security Model

This document defines the security decisions for AUTH v0.

## Dependencies

Required external dependencies are intentionally limited to:

```text
github.com/mattn/go-sqlite3
github.com/go-playground/validator/v10
golang.org/x/crypto/argon2
```

Everything else should prefer the Go standard library.

Relevant standard-library functionality includes:

```text
crypto/rand       random salts and session secrets
crypto/sha256     session-token hashing
crypto/subtle     constant-time comparisons
encoding/base64   encoded salts, hashes, and session tokens
net/http          HTTP server and request handling
uuid              UUIDv4 session IDs
database/sql      database access
strings           normalization
time              session expiration
```

## Passwords

### Algorithm

Passwords use **Argon2id** through `golang.org/x/crypto/argon2.IDKey`.

AUTH must not implement Argon2id itself.

### Initial parameter target

The initial configuration to benchmark is:

```text
memory:       64 MiB (65536 KiB)
time/passes:  3
parallelism:  1
salt:         16 random bytes
output:       32 bytes
```

These parameters are not immutable. They should be benchmarked in the actual AUTH container and adjusted to keep legitimate authentication practical while making offline password guessing expensive.

AUTH uses OWASP's 19 MiB / 2 iterations / parallelism 1 baseline as its minimum configuration:

```text
memory:       19 MiB
time/passes:  2
parallelism:  1
```

### Configuration bounds

AUTH accepts these Argon2id cost settings:

| Environment variable | Default | Minimum | Maximum |
| -------------------- | ------: | ------: | ------: |
| `ARGON2_MEMORY_KIB`  |   65536 |   19456 |  262144 |
| `ARGON2_ITERATIONS`  |       3 |       2 |      10 |
| `ARGON2_PARALLELISM` |       1 |       1 |       4 |

These bounds are project policy. Defaults and upper bounds must be
benchmarked in the AUTH container before deployment.

### Password handling

The exact password received during registration is the value passed to Argon2id. Login verification must use the submitted password exactly as received.

AUTH does not apply Unicode normalization, keeping password processing within the chosen dependency set. Visually identical passwords with different Unicode representations are treated as different passwords.

Passwords are treated as opaque input.

AUTH must not:

```text
trim whitespace
change letter casing
otherwise normalize the password
```

Plaintext passwords must only remain in memory for as long as required to hash or verify them.

### Password policy

For v0, passwords must contain 15 to 128 Unicode code points.
Spaces and Unicode characters are allowed. Passwords are never truncated.

AUTH does not require any particular combination of uppercase letters, lowercase letters, numbers, or symbols.

v0 does not check passwords against common or compromised password blocklists.

Passwords are not trimmed, lowercased, or otherwise normalized. The password is processed exactly as submitted.

FRONTEND may enforce and display the same requirements for user experience, but AUTH remains authoritative for password validation.

### Password storage

Store a self-describing value containing:

```text
algorithm
version
memory cost
time cost
parallelism
salt
hash
```

Example shape:

```text
$argon2id$v=19$m=65536,t=3,p=1$<salt>$<hash>
```

A new cryptographically random salt is generated for every password.

The plaintext password is never stored.

### Verification

On login:

1. Parse the stored Argon2id parameters.
2. Validate the parsed cost parameters against accepted minimum and maximum bounds before passing them to Argon2id.
3. Decode the salt and stored hash.
4. Derive a hash from the submitted password using the stored parameters.
5. Compare the derived and stored hash bytes in constant time.

A malformed password-hash string or unsupported parameter set must fail safely rather than panic or cause excessive resource usage.

## Sessions

Session identity and session authentication are deliberately separate.

```text
session.id     UUIDv4 identifier
session token  32 cryptographically random bytes, secret
```

The UUID identifies the session. It is not an authentication credential.

### Session lifetime

Sessions expire after 7 days of inactivity or 30 days after creation,
whichever comes first.

These durations are configurable through `SESSION_IDLE_TIMEOUT_SECONDS`
(default: 604800) and `SESSION_LIFETIME_SECONDS` (default: 2592000).
Both must be positive integers; zero is invalid and cannot disable either timeout.

`now` means the current server-side Unix timestamp in seconds. `idle_timeout`
is the configured idle timeout in seconds.

AUTH sets `expires_at` at creation and never extends it. Validation rejects
sessions when `expires_at <= now` or
`last_seen_at + idle_timeout <= now`, before recording new activity.

Successful validation updates `last_seen_at`. v0 records activity on every
successful validation.

Configuration changes affect the absolute lifetime of newly created
sessions only. The current idle timeout applies to all sessions.

### Session creation

On successful login:

1. Create a UUIDv4 `sessions.id`.
2. Generate a separate 32-byte session secret using `crypto/rand`.
3. Compute SHA-256 over the raw 32-byte secret.
4. Store only the resulting 32-byte hash in `sessions.token_hash`.
5. Encode the raw secret using unpadded Base64URL for transport.
6. Initialize `last_seen_at` to the creation time and set `expires_at` to the
   creation time plus the configured absolute lifetime (default: 30 days).
7. Return the encoded opaque session token to BACKEND over the private service boundary.

The raw session secret and its Base64URL representation are never stored in SQLite.

### Why SHA-256 is appropriate for session tokens

Passwords have relatively low, human-generated entropy and require an intentionally expensive password-hashing function.

Session secrets are generated with 256 bits of cryptographic randomness. They are not realistically brute-forceable, so a fast one-way hash is appropriate for storage and database lookup.

### Session validation

AUTH must reject a session if:

- the token is missing or malformed;
- Base64URL decoding fails;
- the decoded token is not exactly 32 bytes;
- its SHA-256 hash is not found;
- `expires_at <= now`;
- `last_seen_at + idle_timeout <= now`.

AUTH enforces both absolute expiry and idle timeout using server-side time.
Expired sessions are rejected before `last_seen_at` is updated. Successful
validation records activity without extending `expires_at`.

### Session revocation

Logout is idempotent. AUTH returns `204 No Content` for missing, malformed,
unknown, expired, and active tokens. A matching active or expired session row
is deleted when possible. The response does not reveal token validity,
session existence, or expiration status.

AUTH returns `500 internal_error` only for database or unexpected internal
failures. BACKEND clears the browser cookie regardless of the `204` outcome.

This provides immediate server-side revocation because the session credential no longer resolves to a valid session.

## Cookies

AUTH does not own the browser-facing session cookie. BACKEND does.

The intended secure deployment configuration is:

```text
__Host-session=<opaque-token>
Path=/
HttpOnly
Secure
SameSite=Lax
no Domain attribute
Max-Age/Expires matching the returned absolute expires_at
```

`HttpOnly` prevents normal frontend JavaScript from reading the credential.

`Secure` restricts the cookie to HTTPS transport.

The `__Host-` prefix requires `Secure`, `Path=/`, and no `Domain` attribute.

For local development over plain HTTP, the `Secure` attribute may need to be disabled and the `__Host-` prefix may therefore not be usable. These relaxed settings are development-only.

AUTH enforces both timeouts even if a client continues sending the cookie.
A session can expire through inactivity before its cookie expires.

## Request handling

### Resource limits

Registration and login request bodies must not exceed 8 KiB (8192 bytes).
AUTH enforces this limit while reading the body, regardless of whether
`Content-Length` is present or accurate. Bodies exceeding the limit return
`413 request_too_large` using the standard JSON error envelope.

`ARGON2_MAX_CONCURRENCY` sets the maximum number of simultaneous Argon2id
operations per AUTH process. It defaults to 2 and must be a positive integer.
Registration hashing, login verification, and unknown-email dummy
verification share one non-blocking limit.

When no hashing slot is available, AUTH returns `503 service_busy` without
queuing or starting Argon2id. Slots are released when operations finish,
including failure paths. Session validation and logout do not consume
hashing slots.

At the initial 64 MiB memory setting, two simultaneous operations use about
128 MiB for Argon2id, in addition to other process memory. This limit bounds
concurrent operations, not total process memory. Benchmark the default in
the AUTH container before finalizing deployment configuration.

These limits protect AUTH's resources independently of BACKEND's public
rate limiting.

### Never log

Do not log:

```text
plaintext passwords
registration or login request bodies
session cookie values
opaque session tokens
internal Authorization headers containing session tokens
password hashes
```

Sensitive authentication values must also be excluded or redacted from structured request and response logging.

### Login errors

Login returns the same authentication failure for:

```text
unknown email
incorrect password
```

Example:

```text
invalid email or password
```

This prevents the login endpoint from directly revealing whether a particular email address exists.

When the email is unknown, AUTH performs password verification against a fixed dummy Argon2id hash using the current hashing parameters before returning invalid_credentials. The dummy hash is prepared once and reused, avoiding an immediate rejection that could reveal account existence through timing.

### SQL

Use parameterized queries everywhere.

Never concatenate user-controlled input into SQL statements.

### Uniqueness

Application-level checks may be used where helpful, but SQLite `UNIQUE` constraints remain the authoritative and race-safe mechanism for enforcing unique email addresses and usernames.

## Public-edge responsibility

Because AUTH is private, BACKEND owns protections at the public application boundary, including:

- rate limiting or throttling login and registration;
- request body size limits;
- HTTPS termination when deployed over a network;
- CSRF protection or origin validation for cookie-authenticated state-changing requests;
- secure browser cookie attributes;
- prevention or redaction of sensitive request and response logging.

AUTH still validates authentication-specific input and must not rely on FRONTEND or BACKEND validation for correctness.

Additional defense-in-depth controls may be added later without changing this responsibility boundary.

## References

- OWASP Password Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- NIST SP 800-63B - Passwords: https://pages.nist.gov/800-63-4/sp800-63b/authenticators/#passwords
- Go Argon2 package: https://pkg.go.dev/golang.org/x/crypto/argon2
- Go UUID package: https://pkg.go.dev/uuid
- MDN Set-Cookie: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Set-Cookie
