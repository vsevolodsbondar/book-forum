# User API

Covers the User endpoints: create (register), get, update, delete — and the shared error response format.

## Errors

Every error response uses the same JSON envelope, produced by the shared `GlobalErrorHandler`:

```json
{
  "status": 404,
  "message": "user not found"
}
```

The HTTP status code matches `status` in the body. Sentinel errors map to status codes like this:

| Sentinel error | HTTP status | Meaning |
|---|---|---|
| `ErrUserNotFound` | 404 | The referenced user doesn't exist |
| `ErrForbidden` | 403 | Authenticated, but not the owner of this account |
| `ErrExpiredSession` | 401 | Session cookie missing/invalid/expired |
| `ErrInvalidInput`, `ErrBadRequest` | 400 | Malformed request (invalid `id`, bad JSON, unknown fields)
