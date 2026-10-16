# How to run the project

Step-by-step instructions for starting the Literary Lions forum on a fresh machine, right after `git clone`.

The project consists of three services that are started together with Docker Compose:

| Service | What it is | Address |
|---|---|---|
| `frontend` | nginx with the HTML / JS pages | http://localhost:5173 |
| `backend` | Go REST API + SQLite database | http://localhost:8080 |
| `auth` | Go authentication service (registration, login, sessions) | internal only, reachable from `backend` |

## 1. Requirements

* **Git**
* **Docker Desktop** (or Docker Engine) with **Docker Compose v2**. Check it:

```bash
docker --version
docker compose version
```

Nothing else is needed: Go, Node.js and SQLite are not required on your computer, everything is built inside the containers.

## 2. Clone the repository

```bash
git clone https://gitea.kood.tech/vsevolodbondar/forum.git
cd forum
```

(Gitea may ask for your login and password / access token.)

## 3. Create the settings file of the auth service

The auth service reads its settings from `auth/.env`. This file is **not stored in git**, so it has to be created from the example once:

```bash
cp auth/.env.example auth/.env
```

The default values work as they are, no editing is needed. Without this file `docker compose up` fails with an error about the missing `./auth/.env`.

## 4. Start everything

From the project root (the folder with `docker-compose.yml`):

```bash
docker compose up --build
```

The first start takes a few minutes (the images are built). When it is ready, the logs of the backend contain:

```text
Connected to SQLite
Database seeded
```

Open **http://localhost:5173** in the browser.

> Use exactly `localhost:5173`, not `127.0.0.1` and not another port. The backend allows requests (CORS and cookies) only from the origin `http://localhost:5173`.

To start in the background, add `-d`: `docker compose up --build -d`.

## 5. Sample data

The backend is started with the `--init` flag (see `backend/Dockerfile`), so the database is filled with sample data: 6 users, categories, posts, comments and likes / dislikes.

The auth service ships a matching database with the same 6 users (ids 1–6), so the posts and the accounts belong together:

| id | username | email |
|---|---|---|
| 1 | alice | alice@example.com |
| 2 | bob | bob@example.com |
| 3 | charlie | charlie@example.com |
| 4 | diana | diana@example.com |
| 5 | eve | eve@example.com |
| 6 | frank | frank@example.com |

The passwords of these sample users are not written in the repository (only their hashes are stored). Ask the team, or simply **register your own user** on the login page.

Registration rules:

* email: ASCII characters only;
* username: 3–30 characters, letters, digits and `_`;
* password: 15–128 characters.

## 6. Stop and reset

Stop the project (keeps the data): press `Ctrl+C` in the terminal, or run in another terminal:

```bash
docker compose stop
```

**Remove the containers and reset all data** (both databases are stored inside the containers):

```bash
docker compose down
docker compose up --build
```

After a reset:

* the sample data is created again;
* users you registered are gone.

Always reset **both** databases together (`down` does it): the backend and the auth service keep their own users, and they must match.

**Optional: if something went wrong after a reset.** If you were logged in before the reset, the browser still keeps the old `session_id` cookie, but the session it points to was deleted together with the database. Then the page may open without the header buttons (no avatar, no "Log out") and the browser console shows `GET /users/{id} 404`. To fix it, delete the cookie: DevTools → Application → Cookies → `http://localhost:5173` → delete `session_id`, and reload the page (or just log in again). This is not needed on the first start.

## 7. Useful commands

```bash
docker compose ps                 # which services are running
docker compose logs -f backend    # logs of one service (backend / auth / frontend)
docker compose up --build -d frontend   # rebuild and restart only the frontend
docker compose down               # remove containers (and reset the data)
```

## 8. More documentation

* `docs/frontend.md` — the frontend: pages, scripts, used endpoints
* `docs/posts_api.md`, `docs/comment_api.md`, `docs/users_api.md` — backend API
* `docs/back_auth_integration.md` — how the backend talks to the auth service
* `docs/migration.md` — database migrations
* `auth/docs/` — documentation of the auth service
