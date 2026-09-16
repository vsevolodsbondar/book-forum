# AUTH Database Schema

AUTH owns a dedicated SQLite database.

The initial schema contains two tables:

```text
users 1 ---- * sessions
```

A user may have multiple sessions. Deleting a user also deletes all sessions belonging to that user.

## Source of truth

The SQL migrations in `migrations/` are the authoritative definition of the AUTH database schema.

This document explains the structure and design decisions but does not replace the migrations.

## Migration naming

```text
migrations/
  0001_init.sql
```

Future schema changes receive a new numbered migration rather than modifying migrations that may already have been applied.

## `0001_init.sql`

```sql
CREATE TABLE users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL COLLATE NOCASE UNIQUE,
    password_hash TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token_hash BLOB NOT NULL UNIQUE
        CHECK (length(token_hash) = 32),
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    last_seen_at INTEGER NOT NULL DEFAULT (unixepoch()),
    expires_at INTEGER NOT NULL,

    FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CHECK (expires_at > created_at)
);

CREATE INDEX idx_sessions_user_id
    ON sessions(user_id);

CREATE INDEX idx_sessions_expires_at
    ON sessions(expires_at);
```

## SQLite connection

Foreign-key enforcement is a connection setting and must be enabled by the Go SQLite connection configuration.

For example:

```go
sql.Open("sqlite3", path+"?_foreign_keys=on")
```

The migration defines the foreign-key relationship. The connection configuration enables SQLite to enforce it.

## `users`

### `id`

Stable internal user identifier.

```text
INTEGER PRIMARY KEY AUTOINCREMENT
```

SQLite automatically assigns the value when a user is inserted. Automatically assigned IDs are never reused after deletion, preserving
stable identity references in other services. IDs may contain gaps.

BACKEND may store this ID as `user_id` in application-owned data.

### `email`

Normalized login email.

AUTH normalizes incoming email before insertion or lookup:

```text
trim surrounding whitespace
lowercase
```

The database `UNIQUE` constraint remains the final authority for preventing duplicate email addresses.

### `username`

Display and tagging username.

The original display casing is stored, while `COLLATE NOCASE UNIQUE` prevents ASCII case-only duplicates such as:

```text
Alice
alice
ALICE
```

AUTH trims and validates usernames before storage according to the
[registration username requirements](api.md#username-requirements).

### `password_hash`

Self-describing Argon2id password verifier stored as text.

Conceptual format:

```text
$argon2id$v=19$m=...,t=...,p=...$<salt>$<hash>
```

The encoded value contains the algorithm parameters, salt, and derived password hash.

The plaintext password is never stored.

### Timestamps

`created_at` and `updated_at` are Unix timestamps expressed in seconds.

`created_at` is assigned automatically when the user is inserted.

`updated_at` must be explicitly updated by application SQL whenever the user record changes.

## `sessions`

### `id`

UUIDv4 identifying the login session.

This value identifies and manages the session. It is not the session credential.

### `user_id`

Foreign key referencing `users.id`.

Deleting a user deletes all sessions belonging to that user through `ON DELETE CASCADE`.

### `token_hash`

SHA-256 hash of a separate 32-byte cryptographically random session secret.

```text
browser cookie: Base64URL-encoded session secret
AUTH database:  SHA-256(session secret), stored as a 32-byte BLOB
```

The Base64URL representation is used only to transport the session secret safely as text.

The raw session secret and its Base64URL representation are never stored in SQLite.

### `created_at`

Session creation time expressed as Unix seconds.

### `last_seen_at`

Timestamp of the session's most recently recorded activity.

AUTH initializes this field at session creation and updates it on every
successful validation in v0. Sessions reaching the configured idle timeout
(default: 7 days) are rejected before activity is updated.

### `expires_at`

Absolute server-side session expiration time expressed as Unix seconds.

Set at creation using the configured absolute lifetime (default: 30 days).
This value is never extended by activity.

AUTH rejects a session when `expires_at <= now` or
`last_seen_at + idle_timeout <= now`, even if the browser still sends its cookie.
Here, `now` is the current server-side Unix timestamp in seconds and
`idle_timeout` is the configured idle timeout in seconds.

## Indexes

`idx_sessions_user_id` supports operations that need to find all sessions belonging to a user.

`idx_sessions_expires_at` supports expiration-related queries and future cleanup of expired sessions.

`token_hash` already has a `UNIQUE` constraint, so SQLite creates an index that supports session-token lookup.

## Why TEXT for password hashes and BLOB for session hashes?

`password_hash` contains the algorithm parameters, salt, and derived hash. A self-describing text representation makes those values easy to store together and allows password-hashing parameters to be upgraded later.

`token_hash` is always a fixed 32-byte SHA-256 digest. Storing those raw bytes as a `BLOB` is compact and avoids unnecessary text encoding.

## Schema visualization

The migration SQL can be imported into tools such as drawDB or DBeaver to generate an ER diagram.

The generated diagram is only a visualization. `migrations/0001_init.sql` remains the source of truth for the schema.
