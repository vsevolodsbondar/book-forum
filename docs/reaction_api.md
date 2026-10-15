# Reaction API

Covers the Reaction endpoints: create, update, delete a like/dislike on a comment, and the shared error response format.

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
| `ErrInvalidInput`, `ErrBadRequest` | 400 | Malformed request (bad JSON, unknown fields, body too large) |
| `ErrCommentNotFound` | 404 | The comment being reacted to doesn't exist |
| `ErrReactionNotFound` | 404 | The caller has no reaction on this comment (update/delete) |
| anything else (`ErrCreateReaction`, `ErrUpdateReaction`, `ErrDeleteReaction`, `ErrGetRowsAffected`) | 500 | Unexpected internal failure; logged server-side, message not leaked to the client |

## Data model

A user has **at most one reaction per comment** (`UNIQUE(user_id, comment_id)`). A reaction has a value of `0` or `1` (`CHECK(type_of_like IN (0,1))`), where `1` is a like and `0` is a dislike. Reactions are removed automatically when the comment or the user is deleted (`ON DELETE CASCADE`).

Reaction totals and the caller's own reaction (`likes`, `dislikes`, `is_liked`, `is_disliked`) are returned as part of the comment list, see `docs/comment_api.md`.

## Request body

All endpoints require a valid session and accept the same body:

```json
{
  "comment_id": 5,
  "like": 1
}
```

- `comment_id`: the comment to react to.
- `like`: `1` for a like, `0` for a dislike. Ignored by delete.
- `user_id` comes from the authenticated session, never from the body. If a client sends it anyway, the value is overwritten by the session's user id, so a reaction can't be made on behalf of another user.
- Unknown fields are rejected (`ErrInvalidInput`, 400). The body is limited to 1 MB (`ErrBadRequest`, 400).

> **Note:** a missing `like` is decoded as `0`, so a body without it is treated as a dislike.

## Create a reaction

`POST /reactions`

**Behavior**

- Rejects with `ErrExpiredSession` (401) if the session cookie is missing or the session fails to validate.
- Rejects with `ErrCommentNotFound` (404) if `comment_id` doesn't refer to an existing comment. The same status is returned if the session's user has no local profile row, since both are foreign-key violations.
- **Creating is idempotent:** if the caller already has a reaction on this comment, nothing changes and the request still succeeds. To change an existing reaction use update.

**Response**

`201 Created`, empty body.

## Update a reaction

`PATCH /reactions`

**Behavior**

- Rejects with `ErrExpiredSession` (401) if the session cookie is missing or the session fails to validate.
- Changes the value of the caller's existing reaction on `comment_id` to `like`.
- Rejects with `ErrReactionNotFound` (404) if the caller has no reaction on this comment.

**Response**

`200 OK`, empty body.

## Delete a reaction

`DELETE /reactions`

**Behavior**

- Rejects with `ErrExpiredSession` (401) if the session cookie is missing or the session fails to validate.
- Removes the caller's reaction on `comment_id`. Only `comment_id` is used, `like` is ignored.
- Rejects with `ErrReactionNotFound` (404) if the caller has no reaction on this comment.

**Response**

`204 No Content`.
