# Architecture

## Topology

AUTH is an internal service. Only BACKEND should be reachable from the browser for application API requests.

```mermaid
flowchart LR
    B[Browser] --> F[FRONTEND]
    F -->|/api/*| BE[BACKEND]
    BE -->|private HTTP| A[AUTH]
    A --> AD[(auth.sqlite)]
    BE --> BD[(backend DB)]
```

The browser never receives AUTH's database representation and never talks directly to `auth.sqlite` or internal AUTH endpoints.

## Trust boundaries

### FRONTEND

FRONTEND is untrusted from AUTH's perspective. Client-side validation exists for user experience only and must never be treated as authoritative.

### BACKEND

BACKEND is the public API / BFF. It:

- parses public HTTP requests;
- rejects malformed or oversized requests;
- applies public rate limits;
- forwards only the minimal typed payload required by AUTH;
- owns browser-facing cookie creation and deletion;
- asks AUTH to validate sessions;
- uses the returned `user_id` for application logic.

AUTH remains authoritative for authentication-specific validation.

BACKEND must **not** query AUTH's SQLite database directly.

### AUTH

AUTH is authoritative for:

- email and username validation rules;
- email normalization;
- password rules;
- password hashing and verification;
- user uniqueness and integrity;
- session creation;
- session validation;
- session expiration;
- session revocation.

AUTH validates requests even if BACKEND has already performed transport-level validation.

## Registration sequence

```mermaid
sequenceDiagram
    participant U as Browser / FRONTEND
    participant B as BACKEND
    participant A as AUTH
    participant D as auth.sqlite

    U->>B: POST /api/auth/register<br/>email, username, password
    B->>B: Parse JSON + public HTTP checks
    B->>A: POST /v1/register<br/>email, username, password
    A->>A: Validate + normalize
    A->>A: Argon2id(password)
    A->>D: INSERT user
    D-->>A: Created user / constraint error
    A-->>B: 201 safe user response
    B-->>U: 201 registration result
```

No session is created during registration in v0.

## Login sequence

```mermaid
sequenceDiagram
    participant U as Browser / FRONTEND
    participant B as BACKEND
    participant A as AUTH
    participant D as auth.sqlite

    U->>B: POST /api/auth/login<br/>email, password
    B->>A: POST /v1/login<br/>email, password
    A->>D: SELECT user by normalized email
    D-->>A: password_hash
    A->>A: Verify Argon2id
    A->>A: Generate UUIDv4 session ID
    A->>A: Generate 32-byte random session secret
    A->>A: SHA-256(session secret)
    A->>D: INSERT session
    A-->>B: Session token + expiry + safe identity
    B->>B: Set secure HttpOnly cookie
    B-->>U: 200 login result
```

## Protected request sequence

```mermaid
sequenceDiagram
    participant U as Browser
    participant B as BACKEND
    participant A as AUTH
    participant D as auth.sqlite

    U->>B: GET/POST /api/... + session cookie
    B->>A: POST /v1/session/validate<br/>Bearer opaque-token
    A->>A: Decode token + SHA-256
    A->>D: Lookup token_hash
    D-->>A: Session + user identity
    A->>A: Check absolute expiry + idle timeout
    A->>D: Record last_seen_at for valid session
    D-->>A: Activity recorded
    A-->>B: user_id, username, session_id
    B->>B: Application authorization / business logic
    B-->>U: Application response
```

## Logout sequence

```mermaid
sequenceDiagram
    participant U as Browser
    participant B as BACKEND
    participant A as AUTH
    participant D as auth.sqlite

    U->>B: POST /api/auth/logout + session cookie
    B->>A: POST /v1/logout<br/>Bearer opaque-token
    A->>D: DELETE matching session
    A-->>B: 204 No Content
    B->>B: Expire browser cookie
    B-->>U: 204 No Content
```

Logout is idempotent: AUTH returns indistinguishable `204 No Content`
responses for all client-token outcomes. BACKEND clears the cookie regardless
of the result; an internal failure does not guarantee server-side revocation.

## Data ownership

AUTH owns:

```text
users
sessions
password hashes
session token hashes
```

BACKEND owns application data, for example:

```text
profiles
avatar/profile-picture references
posts
comments
application settings
application permissions
```

BACKEND stores AUTH's `user_id` as the stable identity reference wherever application data needs an owner.
