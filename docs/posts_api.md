# Get Posts

Retrieves a paginated list of posts with optional filtering by author or category and configurable ordering.

## Endpoint

```http
GET /posts
```

## Query Parameters

| Parameter  | Type    | Required | Default | Description                                                                               |
| ---------- | ------- | -------: | ------- | ----------------------------------------------------------------------------------------- |
| `page`     | integer |       No | `1`     | Page number. Must be greater than or equal to 1.                                          |
| `limit`    | integer |       No | `10`    | Number of posts per page. Must be greater than 0.                                         |
| `search`   | string  |       No | —       | Enables search mode when a non-empty value is provided.                                   |
| `field`    | string  |       No | —       | Field used for filtering. Supported values: `author`, `category_id`.                      |
| `value`    | string  |       No | —       | Value used to filter the selected field.                                                  |
| `byLatest` | boolean |       No | `true`  | Controls ordering. `true` returns newest posts first; `false` returns oldest posts first. |

### Pagination

Pagination is calculated using:

```text
offset = (page - 1) * limit
```

For example:

```http
GET /posts?page=2&limit=10
```

returns posts 11–20.

## Search

Search is controlled by the `field` and `value` parameters.

### Search by author

```http
GET /posts?search=true&field=author&value=42
```

Returns posts where:

```text
author_id = 42
```

### Search by category

```http
GET /posts?search=true&field=category_id&value=3
```

Returns posts where:

```text
category_id = 3
```

### Important

When `field` is provided, `value` must also be provided.

Supported fields are:

```text
author
category_id
```

An unsupported field results in a `400 Bad Request`.

## Ordering

By default, posts are returned with the latest posts first.

```http
GET /posts
```

is equivalent to:

```http
GET /posts?byLatest=true
```

To return the oldest posts first:

```http
GET /posts?byLatest=false
```

The ordering is based on the post ID, which is an auto-incrementing value and therefore approximately corresponds to creation order.

## Combining Parameters

Query parameters can be combined.

For example:

```http
GET /posts?page=2&limit=20&search=true&field=category_id&value=5&byLatest=false
```

This requests:

* page `2`
* `20` posts per page
* posts from category `5`
* oldest posts first

## Response

### Status: `200 OK`

```json
{
  "posts": [
    {
      "id": 15,
      "title": "My first post",
      "author_id": 42,
      "parent_comment_id": 3,
      "initial_comment_id": 10,
      "created_at": "2026-09-23T12:30:00Z",
      "commentIDs": [10, 11, 12],
      "likes": 7
    },
    {
      "id": 14,
      "title": "Another post",
      "author_id": 17,
      "parent_comment_id": null,
      "initial_comment_id": 9,
      "created_at": "2026-09-22T18:20:00Z",
      "commentIDs": [9, 13],
      "likes": -2
    }
  ],
  "page": 1,
  "pageSize": 10,
  "totalPages": 5
}
```

### Response fields

#### `posts`

Array containing the posts returned for the requested page.

| Field                | Type         | Description                                          |
| -------------------- | ------------ | ---------------------------------------------------- |
| `id`                 | integer      | Unique post ID.                                      |
| `title`              | string       | Post title.                                          |
| `author_id`          | integer/null | ID of the post author.                               |
| `parent_comment_id`  | integer/null | Category ID according to the current DTO definition. |
| `initial_comment_id` | integer/null | ID of the initial comment associated with the post.  |
| `created_at`         | string       | Post creation timestamp.                             |
| `commentIDs`         | integer[]    | IDs of comments belonging to the post.               |
| `likes`              | integer      | Net like score of the initial comment.               |

The `likes` value is calculated as:

```text
positive likes - negative likes
```

For example:

```text
10 positive likes
3 negative likes

likes = 10 - 3 = 7
```

## Error Responses

### Invalid pagination

```http
GET /posts?page=0&limit=10
```

returns:

```http
400 Bad Request
```

when the pagination parameters are invalid.

### Invalid search field

```http
GET /posts?field=title&value=hello
```

returns:

```http
400 Bad Request
```

because `title` is not currently a supported search field.

### Missing search value

```http
GET /posts?field=author
```

returns:

```http
400 Bad Request
```

because a search value is required when a search field is specified.

### No posts found

If the repository does not find any posts, the service returns the configured `ErrPostNotFound` error, which is handled by the global error handler.

## Example Requests

### Get first page

```http
GET /posts
```

### Get 20 posts

```http
GET /posts?limit=20
```

### Get second page

```http
GET /posts?page=2&limit=10
```

### Get oldest posts first

```http
GET /posts?byLatest=false
```

### Search by author

```http
GET /posts?search=true&field=author&value=42
```

### Search by category

```http
GET /posts?search=true&field=category_id&value=5
```

### Combined search and pagination

```http
GET /posts?page=2&limit=10&search=true&field=category_id&value=5&byLatest=true
```

## Request Flow

The endpoint follows the application architecture:

```text
HTTP Request
     │
     ▼
PostHandler.GetPosts()
     │
     ├── Parse pagination parameters
     ├── Parse query parameters
     │
     ▼
PostService.GetAllPosts()
     │
     ├── Validate SearchPostsDTO
     │
     ▼
PostRepository.GetAll()
     │
     ├── Query posts
     ├── Apply search filter
     ├── Apply pagination
     ├── Retrieve comments
     └── Calculate likes
     │
     ▼
PostsPaginated
     │
     ▼
JSON Response
```

## Notes

* `page` and `limit` are handled by the pagination helper.
* Search validation is performed in `SearchPostsDTO.Validate()`.
* The service layer is responsible for validating the DTO before passing it to the repository.
* The repository is responsible for retrieving posts and their related comments/likes.
* `byLatest` defaults to `true`, so newest posts are returned first.
* The endpoint currently supports searching by `author` and `category_id` only.
