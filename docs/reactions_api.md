# Reaction API

Covers the Reaction endpoints: create, update, delete a like/dislike on a comment — and the shared error response format.

## Errors

Every error response uses the same JSON envelope, produced by the shared `GlobalErrorHandler`:

```json
{
  "status": 404,
  "message": "reaction not found"
}
```

The HTTP status code matches `status` in the body. Sentinel errors map to status codes like this:

| Sentinel error | HTTP status | Meaning |
|---|---|---|
| `ErrExpiredSession` | 401 | Session cookie missing/invalid/expired |
| `ErrInvalidInput`, `ErrBadRequest` | 400 | Malformed request (bad JSON, unknown fields, body too large, invalid `like` value) |
| `ErrCommentNotFound` | 404 | The comment being reacted to doesn't exist (create) |
| `ErrReactionNotFound` | 404 | The caller has no reaction on this comment (update, delete) |
| anything else (`ErrCreateReaction`, `ErrUpdateReaction`, `ErrDeleteReaction`, `ErrGetRowsAffected`, auth service failures) | 500 | Unexpected internal failure; message not leaked to the client |

## Data model

A user has **at most one reaction per comment** (`UNIQUE(user_id, comment_id)`). A reaction has a value of `0` or `1` (`CHECK(type_of_like IN (0,1))`). Reactions are deleted automatically together with their comment or user (`ON DELETE CASCADE`).

Reaction totals and the caller's own reaction are returned as part of the comment list, see `docs/comment_api.md`.

## Request body

All three endpoints require a valid session and use the same body:

```json
{
  "comment_id": 5,
  "like": 1
}
```

- `comment_id` — the comment being reacted to.
- `like` — the reaction value, `0` or `1`. Used by create and update, ignored by delete.
- The reaction's owner is **always the authenticated session's user**. `user_id` is never taken from the body: if a client sends it, the value is overwritten, so a reaction can't be made on behalf of someone else.
- Unknown fields are rejected (`ErrInvalidInput`, 400). The body is limited to 1 MB (`ErrBadRequest`, 400).

> **Note:** a missing `like` is decoded as `0`, and a missing `comment_id` as `0`. A body without `like` is therefore treated as a reaction with value `0`.

## Create a reaction

`POST /reaction`

**Behavior**

- Rejects with `ErrExpiredSession` (401) if the session cookie is missing or the session fails to validate.
- Rejects with `ErrInvalidInput` (400) if `like` is not `0` or `1`.
- Rejects with `ErrCommentNotFound` (404) if `comment_id` doesn't refer to an existing comment. The same status is returned if the session's user has no local profile row, since both cases are foreign-key violations.
- **Creating is idempotent:** if the caller already has a reaction on this comment, nothing changes (the existing value is **not** overwritten) and the request still succeeds. To change an existing reaction use update.

**Response**

`201 Created`, empty body.

## Update a reaction

`PATCH /reaction`

**Behavior**

- Rejects with `ErrExpiredSession` (401) if the session cookie is missing or the session fails to validate.
- Rejects with `ErrInvalidInput` (400) if `like` is not `0` or `1`.
- Sets the value of the caller's existing reaction on `comment_id` to `like`.
- Rejects with `ErrReactionNotFound` (404) if the caller has no reaction on this comment. Updating to the same value succeeds.

**Response**

`200 OK`, empty body.

## Delete a reaction

`DELETE /reaction`

**Behavior**

- Rejects with `ErrExpiredSession` (401) if the session cookie is missing or the session fails to validate.
- Removes the caller's reaction on `comment_id`. Only `comment_id` is used.
- Rejects with `ErrReactionNotFound` (404) if the caller has no reaction on this comment.

**Response**

`204 No Content`.
