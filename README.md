# Literary Lions Forum

A web forum for book readers. Topics are grouped by categories, posts and comments can be liked or disliked, and registered users get extra functionality. The project is split into three services that run together in Docker.

## Features

- Registration, login and logout with server-side sessions that expire
- Posts with categories, ordered newest or oldest first, with pagination
- Filtering posts by category and, for logged in users, by posts they liked or disliked
- Comments, that can be edit and delete
- Likes and dislikes on comments, visible to everyone (a post's score is the score of its first comment)
- User profile pages
- Passwords hashed with Argon2id, sessions stored as hashes
- Whole project runs with one command in Docker, data stored in SQLite

To be added: file / avatar upload, a page with the list of users, real-time messaging.

## Architecture

```text
Browser
   |
   v
FRONTEND  (nginx, :5173)
   |  fetch, credentials: "include"
   v
BACKEND   (Go REST API, :8080)  ---->  backend SQLite (posts, comments, profiles)
   |
   |  private HTTP (internal network)
   v
AUTH      (Go service)          ---->  auth SQLite (users, passwords, sessions)
```

| Service | Responsibility | Reachable from |
|---|---|---|
| `frontend` | HTML / CSS / vanilla JS pages | browser, `http://localhost:5173` |
| `backend` | public API, application logic, sets and clears the session cookie | browser, `http://localhost:8080` |
| `auth` | users, credentials, sessions | `backend` only (internal Docker network) |

The browser never talks to `auth`. Email and passwords exist only in the auth service. The backend stores the auth user id and non-sensitive profile data (username, name, description, profile picture).

## Tech stack

- **Backend / Auth:** Go, SQLite (`mattn/go-sqlite3`), Argon2id (`golang.org/x/crypto/argon2`)
- **Frontend:** plain HTML, CSS and JavaScript (ES modules), served by nginx; Vite for local development
- **Infrastructure:** Docker Compose

## Quick start

Requirements: Git and Docker with Docker Compose v2. Nothing else is needed, everything is built inside the containers.

```bash
git clone https://gitea.kood.tech/vsevolodbondar/forum.git
cd forum
cp auth/.env.example auth/.env
docker compose up --build
```

Open **http://localhost:5173**. Use exactly `localhost:5173`: the backend allows requests and cookies only from this origin.

The first start takes a few minutes while the images are built.

To stop: `Ctrl+C`, or `docker compose stop`. To remove the containers and reset all data: `docker compose down`.

Full instructions, sample users and troubleshooting: [docs/how_to_run.md](docs/how_to_run.md).

### Sample data

The backend starts with the `--init` flag and seeds sample users, categories, posts, comments and reactions. The auth service ships a matching database with the same six users (`alice`, `bob`, `charlie`, `diana`, `eve`, `frank`). Their passwords are not in the repository, so register your own user on the login page.

Registration rules: username 3–30 characters (letters, digits, `_`), password 15–128 characters, email ASCII only.

## Project structure

```text
.
├── docker-compose.yml
├── auth/                  # authentication service
│   ├── main.go
│   ├── Dockerfile
│   ├── .env.example       # copy to .env before the first start
│   ├── migrations/
│   ├── internal/          # config, db, handler, model, repository, server, service
│   └── docs/              # auth service documentation
├── backend/               # REST API
│   ├── main.go
│   ├── Dockerfile
│   ├── client/            # HTTP client for the auth service (and a mock)
│   ├── custom_err/        # sentinel errors and the global error handler
│   ├── db/                # SQLite connection and seeding
│   ├── handler/           # HTTP handlers
│   ├── helper/
│   ├── middleware/        # CORS, timeout, panic recovery
│   ├── migrations/        # SQL schema, embedded into the binary
│   ├── model/
│   ├── repository/        # SQL queries
│   ├── server/            # dependency wiring and routes
│   └── service/           # business logic
├── frontend/              # pages and scripts
└── docs/                  # backend and project documentation
```

The backend is layered: handler → service → repository → SQLite. Every error goes through one `GlobalErrorHandler` and is returned as the same JSON envelope:

```json
{ "status": 404, "message": "user not found" }
```

## API overview

| Area | Endpoints | Details |
|---|---|---|
| Users and sessions | `POST /users`, `GET /users`, `GET /users/{id}`, `PATCH /users/{id}`, `DELETE /users/{id}`, `POST /login`, `POST /logout` | [users_api.md](docs/users_api.md) |
| Posts | `GET /posts`, `POST /posts`, `DELETE /posts/{id}`, `GET /categories` | [posts_api.md](docs/posts_api.md) |
| Comments | `GET /posts/{id}/comments`, `POST /posts/{id}/comments`, `GET /comments/{id}`, `PATCH /comments/{id}`, `DELETE /comments/{id}` | [comment_api.md](docs/comment_api.md) |
| Reactions | `POST /reaction`, `PATCH /reaction`, `DELETE /reaction` | [reactions_api.md](docs/reactions_api.md) |
| Current user | `GET /me` | [frontend.md](docs/frontend.md) |

Reading is public: a guest can see posts, comments and profiles. Creating or changing anything requires a session, and users can edit or delete only their own content.

## Authentication and security

- `POST /login` asks the auth service to verify the credentials and create a session. The backend puts the session token into the `session_id` cookie: `HttpOnly`, `SameSite=Lax`, `Path=/`. The token is never returned in JSON, so frontend JavaScript cannot read it.
- For every protected request the backend validates the cookie through the auth service. The owner of a reaction, comment or profile always comes from the session, never from the request body.
- Passwords are hashed with Argon2id. Session tokens are 32 random bytes; only their SHA-256 hash is stored.
- A session expires after 7 days of inactivity or 30 days after creation, whichever comes first.
- The auth service is on an internal Docker network and has no published port.
- CORS (with credentials) is allowed only for `http://localhost:5173`.

More: [back_auth_integration.md](docs/back_auth_integration.md), [auth/docs/security.md](auth/docs/security.md).

## Database

There are two SQLite databases, one per service, and they must be reset together (`docker compose down` does it). The backend schema is created automatically on startup from the embedded migrations: see [docs/migration.md](docs/migration.md). The auth schema is described in [auth/docs/schema.md](auth/docs/schema.md).

## Documentation

| Document | What it covers |
|---|---|
| [docs/how_to_run.md](docs/how_to_run.md) | Start, stop and reset the project |
| [docs/frontend.md](docs/frontend.md) | Pages, scripts, endpoints used, conventions |
| [docs/users_api.md](docs/users_api.md) | Users, login, logout |
| [docs/posts_api.md](docs/posts_api.md) | Posts and filtering |
| [docs/comment_api.md](docs/comment_api.md) | Comments |
| [docs/reactions_api.md](docs/reactions_api.md) | Likes and dislikes |
| [docs/back_auth_integration.md](docs/back_auth_integration.md) | How the backend talks to the auth service |
| [docs/migration.md](docs/migration.md) | Backend database migrations |
| [auth/docs/README.md](auth/docs/README.md) | Auth service overview |
| [auth/docs/architecture.md](auth/docs/architecture.md) | Service boundaries and sequences |
| [auth/docs/api.md](auth/docs/api.md) | Internal auth API |
| [auth/docs/integration-contract.md](auth/docs/integration-contract.md) | Contract between auth, backend and frontend |
| [auth/docs/schema.md](auth/docs/schema.md) | Auth database schema |
| [auth/docs/security.md](auth/docs/security.md) | Passwords, sessions, cookies |

## Development

Rebuild only the frontend after a change:

```bash
docker compose up --build -d frontend
```

and reload the page without cache (Ctrl+Shift+R / Cmd+Shift+R).

Frontend with Vite:

```bash
cd frontend
npm install
npm run dev
```

Useful commands:

```bash
docker compose ps                  # running services
docker compose logs -f backend     # logs of one service
```

## License

See [LICENSE](LICENSE).
