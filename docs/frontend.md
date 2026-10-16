# Frontend

## Overview

The frontend of the Literary Lions forum is a multi-page application written in plain HTML, CSS and
vanilla JavaScript (ES modules). There is no framework: every page is a separate `.html` file that loads
its own script. Data comes from the backend REST API (`http://localhost:8080`), pages are rendered in
the browser with `fetch` + `insertAdjacentHTML`.

In Docker the pages are served by **nginx** (port `5173`). For local development it is also possible
to use **Vite** (`npm run dev`).

```text
Browser (HTML + JS modules)
        |  fetch, credentials: "include"
        v
Backend :8080  ->  Auth service (internal network)
```

## Project structure

```text
frontend/
├── index.html          # all posts
├── postInside.html     # one post with its comments
├── createPost.html     # form for a new post
├── myPosts.html        # posts of a user
├── login.html          # login and registration forms (one page, two forms)
├── profile.html        # user profile (`?id=` or the logged in user)
├── users.html          # list of users (not implemented yet)
├── Dockerfile          # nginx:alpine, copies the folder to /usr/share/nginx/html
├── biome.json          # formatter / linter settings
└── src/static/
    ├── scripts/        # one JS module per page + header.js shared by all pages
    ├── styles/main.css # all styles
    ├── imgs/           # logo, default avatar, icons
    └── fonts/
```

## Pages and scripts

| Page | Script | What it does |
|---|---|---|
| every page | `header.js` | Asks `/me` who is logged in, renders the header (guest: Register / Login; user: Write a post, avatar linked to `profile.html?id=`, Log out), adds "Create post" and "My posts" to the menu (only if the menu does not have them yet), handles **logout**. Exports `getUserID()` used by other scripts. |
| `index.html` | `allPosts.js` | Renders post cards. Filters above the list: **Newest** / **Oldest** (sorting), category `<select>` (filled from `/categories`), and, for logged in users only, **Liked posts** / **Disliked posts**. All filters are done by the backend. **See more** button loads the next page and appends it. See "Filters on the posts page" below. |
| `postInside.html` | `postInside.js` | Renders the post and its comments, creates / edits / deletes a comment, deletes the post (only for the author). Like / dislike buttons work (see "Likes and dislikes" below). A guest can read the post and the comments. |
| `createPost.html` | `createPost.js` | Loads categories, sends a new post, redirects to `postInside.html?id=<new id>`. Highlights "Create post" in the menu. |
| `myPosts.html` | `myPosts.js` | Posts of a user (`?id=` or the logged in user), same cards as the all posts page. Highlights "My posts" in the menu. |
| `login.html` | `login.js` | Switches between the login and the registration form, sends login / registration, shows error messages. After a successful login the user goes to `index.html`; after a successful registration the page reloads (`login.html`) and opens on the login form. |
| `profile.html` | `profile.js` | Profile of a user: `?id=` or the logged in user. Texts are inserted with `textContent`. Shows "User not found" for 404. |
| `users.html` | `users.js` | Placeholder. |

## Backend endpoints used

| Method and path | Used by | Notes |
|---|---|---|
| `GET /me` | `header.js` | Returns the id of the logged in user, 401 for a guest (expected). |
| `GET /users/{id}` | `header.js`, `profile.js` | Data for the header and for the profile page. 404 if the user does not exist in the backend database. |
| `POST /users` | `login.js` | Registration. Body: `email`, `password`, `username`, `description`. |
| `POST /login` | `login.js` | Body: `email`, `password`. The server sets the `session_id` cookie. |
| `POST /logout` | `header.js` | Server deletes the session and clears the cookie (204). |
| `GET /posts` | `allPosts.js`, `myPosts.js` | Query: `byLatest=true\|false`, `page`, `search=true&field=categoryId\|authorId\|likedBy\|dislikedBy&value=...`. Response: `posts`, `page`, `pageSize`, `totalPages`. 404 means "no posts". |
| `GET /posts/{id}/comments` | `postInside.js` | Post with all comments. The session is optional: a guest gets the same data with `is_liked` / `is_disliked` equal to `false`. 404 if the post does not exist. |
| `POST /posts` | `createPost.js` | Body: `title`, `category_id`, `init_comment_text`. |
| `DELETE /posts/{id}` | `postInside.js` | Only for the author. |
| `POST /posts/{id}/comments` | `postInside.js` | Create a comment. |
| `GET /comments/{id}` | `postInside.js` | Used to render a created or updated comment, and after deleting a comment to check whether it was removed (404 is expected and means "removed"; the browser prints it in red in the console). |
| `PATCH /comments/{id}` | `postInside.js` | Edit a comment. |
| `DELETE /comments/{id}` | `postInside.js` | Delete a comment (or turn it into "Deleted message" if it has replies). |
| `GET /categories` | `allPosts.js`, `createPost.js` | Items: `{ id, category_name }`. |
| `POST /reaction` | `postInside.js` | Create a like or dislike. Body: `comment_id`, `like` (`1` or `-1`). 201 without a body. |
| `PATCH /reaction` | `postInside.js` | Change the reaction to the opposite one. Same body. 200 without a body. |
| `DELETE /reaction` | `postInside.js` | Remove the reaction. Same body. 204. |

## Filters on the posts page

`allPosts.js` keeps the state of the filters in variables (`byLatest`, `reactionField`) and builds the query in `buildQuery(page)`, so the same state is used by the filters, by the category select and by **See more**.

| Control | Query sent to `GET /posts` | Who sees it |
|---|---|---|
| Newest (default) | `byLatest=true` | everyone |
| Oldest | `byLatest=false` | everyone |
| Category select | `search=true&field=categoryId&value=<id>` | everyone |
| Liked posts | `search=true&field=likedBy&value=<user id>` | logged in users only |
| Disliked posts | `search=true&field=dislikedBy&value=<user id>` | logged in users only |

* The backend supports only one `search` field at a time, so a category and "Liked / Disliked posts" cannot be combined: choosing one resets the other.
* "Liked posts" are the posts whose **first comment (the post text)** the user liked; "Disliked posts" work the same way with a dislike.
* The "Liked / Disliked posts" buttons are hidden by default and shown by `allPosts.js` when `getUserID()` returns an id.

## Likes and dislikes

* A reaction belongs to a comment (the post text is the first comment). Values: **like = `1`, dislike = `-1`**. A user has at most one reaction per comment.
* The buttons `.likes` / `.dislikes` are rendered with the class `choosed` when the current user has already reacted (`is_liked` / `is_disliked` from the backend).
* One click handler is attached to `document` (event delegation, because comments are rendered later). The request depends on the current state of the buttons of that comment:

| State before the click | Clicked | Request | Result in the page |
|---|---|---|---|
| no reaction | like / dislike | `POST /reaction` | button gets `choosed`, its counter +1 |
| the same button is `choosed` | the same button | `DELETE /reaction` | `choosed` removed, counter -1 |
| the opposite button is `choosed` | the other button | `PATCH /reaction` | `choosed` moves, counter -1 on the old button, +1 on the new one |

* The backend answers without a body, so the counters are updated locally after a successful response.
* A guest (401 / 403) is redirected to `login.html`.

## Menu highlight

The menu items "Create post" and "My posts" are added by `header.js` after the `/me` request, so they do not exist when `createPost.js` / `myPosts.js` start. These scripts wait for the item with a `MutationObserver` and add the class `choosed` to its `<li>` when it appears. "Posts" is highlighted in `index.html` directly.

## Sessions and cookies

* Every request that needs the user uses `credentials: "include"`, otherwise the browser does not send the cookie.
* The cookie is `HttpOnly`, so JavaScript cannot read it. The frontend only knows who is logged in by asking `/me`.
* The backend CORS allows credentials only for the origin `http://localhost:5173`.
* A guest sees all posts and comments, but the comment form is disabled and the owner-only buttons are not rendered.

## Error handling

Forms show the message inside a `<p class="form_error">`. The message depends on the HTTP status:

| Form | Status | Message |
|---|---|---|
| login | 401 | Wrong email or password |
| registration | 409 | This email or username is already in use |
| registration | 400 | Check your data (password >= 15 characters, username 3-30 characters: letters, digits, `_`) |
| registration | passwords differ | Passwords do not match (checked in the browser before the request) |
| create post | 401 / 403 | Please log in to create a post |
| create post | 400 | Check the title, the category and the text |
| any | `fetch` failed | The server is not available |
| any | other | Something went wrong |

## Conventions

* One script per page, all loaded with `type="module"`; shared code is imported (`import { getUserID } from "./header.js"`).
* Event delegation is used for buttons that are rendered later by JavaScript (logout, edit / delete comment).
* Buttons that look the same have different classes if they do different things (`delete-post` vs `btn_delete`), otherwise one click can trigger two handlers.
* The `hidden` attribute is overridden by `display` in CSS, so elements are hidden with `style.display`.
* Formatting: Biome (`biome.json`).

## Running

Docker (recommended, run from the project root):

```bash
docker compose up --watch --build
```

Open `http://localhost:5173`. `--watch` copies changed frontend files into the container
(`docker-compose.override.yml`). If the browser shows an old script, rebuild only the frontend:

```bash
docker compose up --build -d frontend
```

and reload the page without cache (Cmd+Shift+R).

Local development with Vite:

```bash
cd frontend
npm install
npm run dev
```

## Known limitations and TODO

* `users.html` (list of users) is a placeholder; `users.js` has no code yet.
* The profile page shows only the username, name and description; there is no button to edit the profile yet.
* Only one search filter works at a time (see "Filters on the posts page").
* Avatar upload is not connected: the file field exists, but the backend expects a string in `profilepic`.
* "Popular" and "Most comments" sorting is not possible: the backend sorts only by date (`byLatest`).
* Post cards show the category name; the filter is by category only (no search yet).
* The links `login.html#reg` / `#login` in the header are not handled, the page always opens on the login form.
* The API address `http://localhost:8080` is written directly in the scripts; it should be moved into one constant.
* User text is inserted with `innerHTML` in templates (titles, comments, usernames). It must be escaped or inserted with `textContent` to avoid XSS.
* After a successful registration the page just reloads on the login form; there is no "You are registered" message.
* If the database is recreated, an old `session_id` cookie in the browser can point to a user that does not exist any more (`GET /users/{id}` returns 404 and the header stays empty). Delete the cookie or log in again.
