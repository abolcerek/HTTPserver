# Chirpy

A RESTful HTTP server built in Go that powers Chirpy, a Twitter-like microblogging service. Users can register, authenticate, and post short messages ("chirps") capped at 140 characters, with full JWT-based authentication, refresh tokens, and webhook support for premium upgrades.

## Features

- **User management** — registration, login, and password updates with bcrypt-hashed passwords
- **JWT authentication** — short-lived access tokens with long-lived refresh tokens (60-day expiry)
- **Token revocation** — endpoint to invalidate refresh tokens on logout
- **Chirps CRUD** — create, read, list, and delete short messages
- **Filtering & sorting** — query chirps by author and sort ascending or descending by creation time
- **Profanity filter** — automatically replaces banned words (`kerfuffle`, `sharbert`, `fornax`) with `****`
- **Premium upgrades** — Polka webhook integration for upgrading users to Chirpy Red
- **Admin tooling** — fileserver hit metrics and dev-only database reset
- **Static file serving** — built-in fileserver at `/app/`
- **Health checks** — `/api/healthz` endpoint for liveness probes

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL
- **Query generation:** [sqlc](https://sqlc.dev/)
- **Authentication:** JWT (HS256) + bcrypt password hashing
- **Routing:** Go's standard `net/http` ServeMux (Go 1.22+ pattern matching)
- **Environment:** [godotenv](https://github.com/joho/godotenv)

## Prerequisites

- Go 1.22 or higher
- PostgreSQL 14 or higher
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) (optional, only if regenerating query code)
- [goose](https://github.com/pressly/goose) or another migration tool (optional, if applying SQL migrations)

## Installation

Clone the repository and install dependencies:

```bash
git clone https://github.com/abolcerek/HTTPserver.git
cd HTTPserver
go mod download
```

## Configuration

Create a `.env` file in the project root with the following variables:

```env
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
PLATFORM=dev
JWT_SECRET=your-256-bit-secret-here
POLKA_KEY=your-polka-api-key-here
```

| Variable     | Description                                                                 |
|--------------|-----------------------------------------------------------------------------|
| `DB_URL`     | PostgreSQL connection string                                                |
| `PLATFORM`   | Set to `dev` to enable the `/admin/reset` endpoint; any other value disables it |
| `JWT_SECRET` | Secret key used to sign and validate JWTs                                   |
| `POLKA_KEY`  | API key required by the Polka webhook for premium upgrades                  |

## Database Setup

Create the database and apply the migrations from the `sql/` directory:

```bash
createdb chirpy
goose -dir sql/schema postgres "$DB_URL" up
```

If you modify any `.sql` query files, regenerate the database layer with:

```bash
sqlc generate
```

## Running the Server

```bash
go run main.go
```

Or build and run the binary:

```bash
go build -o HTTPserver
./HTTPserver
```

The server listens on **port 8080**.

## API Reference

### Public

| Method | Endpoint               | Description                                        |
|--------|------------------------|----------------------------------------------------|
| GET    | `/api/healthz`         | Liveness check, returns `200 OK`                   |
| POST   | `/api/users`           | Register a new user                                |
| POST   | `/api/login`           | Authenticate and receive JWT + refresh token       |
| POST   | `/api/refresh`         | Exchange a refresh token for a new JWT             |
| POST   | `/api/revoke`          | Revoke a refresh token                             |
| GET    | `/api/chirps`          | List all chirps (supports `?author_id=` and `?sort=asc\|desc`) |
| GET    | `/api/chirps/{chirpID}`| Fetch a single chirp by ID                         |

### Authenticated (Bearer JWT required)

| Method | Endpoint                | Description                                        |
|--------|-------------------------|----------------------------------------------------|
| PUT    | `/api/users`            | Update the authenticated user's email and password |
| POST   | `/api/chirps`           | Create a new chirp (max 140 characters)            |
| DELETE | `/api/chirps/{chirpID}` | Delete a chirp owned by the authenticated user     |

### Webhook

| Method | Endpoint               | Description                                                              |
|--------|------------------------|--------------------------------------------------------------------------|
| POST   | `/api/polka/webhooks`  | Polka webhook endpoint; upgrades a user to Chirpy Red on `user.upgraded` events. Requires API key in `Authorization` header |

### Admin

| Method | Endpoint          | Description                                                       |
|--------|-------------------|-------------------------------------------------------------------|
| GET    | `/admin/metrics`  | HTML page showing total fileserver hits                           |
| POST   | `/admin/reset`    | Resets fileserver hits and wipes all users (dev environment only) |

### Static Files

| Path     | Description                                       |
|----------|---------------------------------------------------|
| `/app/*` | Static fileserver rooted at the project directory |

## Example Usage

**Register a new user:**

```bash
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"hunter2"}'
```

**Log in:**

```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"hunter2"}'
```

**Post a chirp:**

```bash
curl -X POST http://localhost:8080/api/chirps \
  -H "Authorization: Bearer <JWT>" \
  -H "Content-Type: application/json" \
  -d '{"body":"Hello, Chirpy!"}'
```

**List chirps by an author, newest first:**

```bash
curl "http://localhost:8080/api/chirps?author_id=<UUID>&sort=desc"
```

## Project Structure

```
HTTPserver/
├── assets/           # Static assets
├── internal/
│   ├── auth/         # JWT, bcrypt, and bearer-token helpers
│   └── database/     # sqlc-generated database queries
├── sql/
│   ├── schema/       # Database migrations
│   └── queries/      # SQL queries consumed by sqlc
├── index.html        # Landing page served at /app/
├── main.go           # Server entry point and HTTP handlers
├── sqlc.yaml         # sqlc configuration
├── go.mod
└── go.sum
```

## Notes

- Chirp bodies must be **strictly fewer than 140 characters**.
- Profanity filtering is case-insensitive but only matches whole, space-delimited words.
- The `/admin/reset` endpoint returns `403 Forbidden` unless `PLATFORM=dev`.
- Refresh tokens expire 60 days after issuance and can be revoked via `/api/revoke`.

## License

Add your license of choice here (e.g., MIT, Apache 2.0).