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
├── profile.html        # user profile (not implemented yet)
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
| every page | `header.js` | Asks `/me` who is logged in, renders the header (guest: Register / Login; user: Write a post, avatar, Log out), adds "Create post" and "My posts" to the menu, handles **logout**. Exports `getUserID()` used by other scripts. |
| `index.html` | `allPosts.js` | Renders post cards. Filters: **New** (newest first), category `<select>` (filled from `/categories`, filters on the backend). **See more** button loads the next page and appends it. |
| `postInside.html` | `postInside.js` | Renders the post and its comments, creates / edits / deletes a comment, deletes the post (only for the author). Like / dislike state is rendered, clicking is not implemented yet. |
| `createPost.html` | `createPost.js` | Loads categories, sends a new post, redirects to `postInside.html?id=<new id>`. |
| `myPosts.html` | `myPosts.js` | Posts of a user (`?id=` or the logged in user), same cards as the all posts page. |
| `login.html` | `login.js` | Switches between the login and the registration form, sends login / registration, shows error messages. |
| `profile.html` | `profile.js` | Placeholder. |
| `users.html` | `users.js` | Placeholder. |

## Backend endpoints used

| Method and path | Used by | Notes |
|---|---|---|
| `GET /me` | `header.js` | Returns the id of the logged in user, 401 for a guest (expected). |
| `GET /users/{id}` | `header.js` | Data for the header. |
| `POST /users` | `login.js` | Registration. Body: `email`, `password`, `username`, `description`. |
| `POST /login` | `login.js` | Body: `email`, `password`. The server sets the `session_id` cookie. |
| `POST /logout` | `header.js` | Server deletes the session and clears the cookie (204). |
| `GET /posts` | `allPosts.js`, `myPosts.js` | Query: `byLatest=true`, `page`, `search=true&field=categoryId\|authorId&value=...`. Response: `posts`, `page`, `pageSize`, `totalPages`. 404 means "no posts". |
| `GET /posts/{id}/comments` | `postInside.js` | Post with all comments. |
| `POST /posts` | `createPost.js` | Body: `title`, `category_id`, `init_comment_text`. |
| `DELETE /posts/{id}` | `postInside.js` | Only for the author. |
| `POST /posts/{id}/comments` | `postInside.js` | Create a comment. |
| `GET /comments/{id}` | `postInside.js` | Used to render a created or updated comment. |
| `PATCH /comments/{id}` | `postInside.js` | Edit a comment. |
| `DELETE /comments/{id}` | `postInside.js` | Delete a comment (or turn it into "Deleted message" if it has replies). |
| `GET /categories` | `allPosts.js`, `createPost.js` | Items: `{ id, category_name }`. |

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

* Like / dislike buttons are rendered but clicking them does nothing (the reaction endpoints are not ready).
* `profile.html` and `users.html` are placeholders.
* Avatar upload is not connected: the file field exists, but the backend expects a string in `profilepic`.
* "Popular" and "Most comments" filters need a `sort` parameter on the backend.
* Post cards show the category name; the filter is by category only (no search yet).
* The links `login.html#reg` / `#login` in the header are not handled, the page always opens on the login form.
* The API address `http://localhost:8080` is written directly in the scripts; it should be moved into one constant.
* User text is inserted with `innerHTML` in templates (titles, comments, usernames). It must be escaped or inserted with `textContent` to avoid XSS.
* After a successful registration the user should see a message and the login form (the success message is not added yet).
