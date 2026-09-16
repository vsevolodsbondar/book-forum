# Authentication Service Documentation

This directory defines the scope, architecture, internal API, security model, integration contract, database schema, and implementation roadmap for the **AUTH** service.

The project has three runtime containers:

```text
FRONTEND  ->  BACKEND  ->  AUTH
                |           |
                |           +-> auth.sqlite
                +-> application data / application logic
```

`AUTH` is a private service. The browser does not call it directly. `BACKEND` is the public application API / BFF and is the only application service expected to call `AUTH`.

## What are we building?

A small, reusable authentication service responsible for:

- registering users with a unique email, unique username, and password;
- verifying email + password logins;
- securely hashing passwords with **Argon2id**;
- creating and managing server-side login sessions;
- using **UUIDv4** identifiers for sessions;
- issuing opaque, random session credentials;
- validating and revoking sessions;
- persisting authentication data in SQLite.

The service deliberately does **not** own application profiles, posts, avatars, business data, or application authorization rules.

## Why is AUTH separate?

AUTH owns one narrow security boundary: **identity and authentication**.

This gives the project a clear separation of responsibilities:

| Component | Owns                                                                   |
| --------- | ---------------------------------------------------------------------- |
| FRONTEND  | UI, forms, client-side UX validation                                   |
| BACKEND   | Public API, browser cookies, app/business logic, authorization, app DB |
| AUTH      | Users, passwords, login verification, sessions, auth DB                |

Keeping AUTH private reduces unnecessary network exposure and prevents BACKEND or FRONTEND from coupling directly to authentication storage.

## How does it work?

### Registration

```text
Browser
  -> BACKEND POST /api/auth/register
  -> AUTH POST /v1/register
  -> validate + normalize input
  -> Argon2id(password)
  -> INSERT users
  <- safe user result
  <- 201 Created
```

Registration creates the account only. It does **not** create a login session in v0.

### Login

```text
Browser
  -> BACKEND POST /api/auth/login
  -> AUTH POST /v1/login
  -> verify Argon2id password
  -> create UUIDv4 session ID
  -> generate 32-byte random session secret
  -> store SHA-256(secret) in SQLite
  <- raw secret + expiry returned internally to BACKEND
  -> BACKEND sets HttpOnly browser cookie
```

The raw session secret is never persisted by AUTH. Only its SHA-256 hash is stored.

### Authenticated request

```text
Browser
  -> BACKEND + session cookie
  -> BACKEND extracts opaque session token
  -> AUTH POST /v1/session/validate
  -> SHA-256(token)
  -> lookup active session
  <- user_id + session identity
  -> BACKEND performs application authorization/business logic
```

### Logout

```text
Browser
  -> BACKEND POST /api/auth/logout
  -> AUTH POST /v1/logout
  -> delete matching session
  <- 204 No Content
  -> BACKEND expires browser cookie
```

## What are we building first?

The first release is deliberately small:

1. SQLite schema and migrations.
2. Argon2id password package.
3. Registration.
4. Login + session creation.
5. Session validation.
6. Logout / revocation.
7. Integration tests and handoff to BACKEND/FRONTEND.

Everything else is post-v0 unless required by the assignment.

## Documents

- [architecture.md](architecture.md) — service boundaries and request timelines.
- [api.md](api.md) — internal AUTH HTTP contract.
- [integration-contract.md](integration-contract.md) — handoff contract for BACKEND and FRONTEND.
- [schema.md](schema.md) — SQLite schema and data model.
- [security.md](security.md) — passwords, sessions, cookies, trust boundaries.
- [roadmap.md](roadmap.md) — implementation order and definition of done.

## v0 non-goals

The following are intentionally outside the first release:

- password reset;
- email verification;
- MFA / TOTP / passkeys;
- OAuth / OIDC;
- JWT access tokens;
- refresh tokens;
- roles and permissions;
- admin APIs;
- audit logging;
- device/session management APIs.

These features can be added later around the core users -> sessions model without changing the basic authentication flow.
