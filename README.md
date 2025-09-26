# Full-Stack PostgreSQL: Go API + React + Docker

A minimal full-stack starter that runs entirely in Docker:
- **Backend:** Go 1.23 (chi + pgx over `database/sql`)
- **DB:** PostgreSQL 17
- **Frontend:** React 18 + Vite + TypeScript
- Ops: Docker Compose, Prometheus metrics, basic rate-limit, Swagger/OpenAPI

---

## Stack

- Go 1.23, chi, pgx
- PostgreSQL 17 (alpine)
- React 18, Vite, TypeScript
- Docker, Docker Compose
- Optional IDEs: GoLand (API), WebStorm/VS Code (web), DBeaver CE (DB client)

---

## Project Layout

```bash
fullstack-postgres/
├─ go-api/
│  ├─ cmd/
│  │  ├─ api/       # HTTP API (main)
│  │  └─ auditor/   # (optional) Kafka auditor (future)
│  ├─ internal/
│  │  ├─ audit/           # (future) consumer
│  │  ├─ cache/           # Redis helpers (JWT revoke, json cache)
│  │  ├─ config/          # env → Config
│  │  ├─ db/              # sql.DB open/pool
│  │  ├─ domain/          # DTOs / models (Item, User)
│  │  ├─ events/          # Kafka writer (optional)
│  │  ├─ http/            # handlers, router, middlewares, openapi.yaml
│  │  ├─ metrics/         # Prometheus registry + middleware
│  │  ├─ migrate/         # migrations runner + sql files
│  │  ├─ repo/            # repositories (pg)
│  │  └─ service/         # business services
│  ├─ .env                # API env (sample below)
│  ├─ Dockerfile
│  ├─ go.mod
│  └─ openapi.yaml        # served at /openapi.yaml and /docs
│
├─ infra/
│  ├─ docker-compose.yml  # all services
│  ├─ logging/            # filebeat/logstash (optional)
│  ├─ postgres/init/      # init + seed SQL
│  ├─ prometheus.yml      # (future) scrape config
│  └─ .env                # compose vars (GOAPI_PATH/WEBAPP_PATH etc.)
│
└─ webapp/                # React app (can live elsewhere, see compose vars)
   ├─ src/
   ├─ public/
   ├─ .env                # VITE_API_URL
   └─ Dockerfile
``` 
Note: You can keep webapp/ outside this repo. Compose accepts absolute paths via infra/.env (see Split folders & paths).

## Quick Start (everything in Docker)
```bash
cd infra
# (optional) set paths if your webapp is outside this repo
cp .env.example .env
# edit .env -> GOAPI_PATH, WEBAPP_PATH, REDIS_PASSWORD, etc.
docker compose up -d --build
```

### Services (defaults):

- API: http://localhost:8080/health → {"ok":true}
- Swagger UI: http://localhost:8080/docs (spec at /openapi.yaml)
- Web: http://localhost:3001 (Nginx serving built Vite app)
- Postgres: localhost:5432 (user/pass: postgres/postgres, db: postgres)
- Redis (optional): localhost:6379 (password from REDIS_PASSWORD)
- Kafka/ELK (optional): exposed as in compose; can be disabled if not needed.

---

## Environment

```bash
PORT=8080

# Either fill individual DB_* or use DB_URL; DB_URL wins if set
DB_HOST=postgres
DB_PORT=5432
DB_NAME=postgres
DB_USER=postgres
DB_PASS=postgres
SSL_MODE=disable
DB_URL=postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable

CORS_ORIGINS=http://localhost,http://localhost:3000,http://localhost:3001

# Rate limit (token bucket, in-process)
RATE_LIMIT_RPS=5
RATE_LIMIT_BURST=10

# JWT (dev defaults)
JWT_ACCESS_SECRET=change-this
JWT_REFRESH_SECRET=change-this-too
JWT_ACCESS_TTL_MIN=15
JWT_REFRESH_TTL_DAYS=7

# Sentry (optional)
SENTRY_DSN=
SENTRY_ENV=dev

# Redis (optional; for JWT revoke + JSON cache)
REDIS_ADDR=redis:6379
REDIS_PASSWORD=change-me
```

#### webapp/.env :

```bash
VITE_API_URL=http://localhost:8080
```
#### infra/.env :
```bash
# Absolute or relative paths to your projects
GOAPI_PATH=../go-api
WEBAPP_PATH=../webapp

# Optional infra secrets
REDIS_PASSWORD=change-me
```

## Database & Seed

#### Init scripts run automatically on first boot (infra/postgres/init):

- 01-create-user.sql (demo users: admin/user)
- 02-schema.sql (tables, constraints)
- 99_seed.sql / 99_seed_items.sql (bulk items; categories, stock)

Re-run item seeding manually:
```bash
docker compose exec postgres psql -U postgres -d postgres \
  -v "ON_ERROR_STOP=1" -f /docker-entrypoint-initdb.d/99_seed_items.sql
```
Check:
```bash
docker compose exec postgres psql -U postgres -d postgres -c "SELECT COUNT(*) FROM app.items;"
```
## API Overview

- Health: GET /health

- Auth

  - POST /auth/login → {access_token, refresh_token}

  - POST /auth/refresh (header X-Refresh-Token)

  - GET /auth/me (Bearer)

- Items (Bearer: roles user or admin; DELETE requires admin)

  - GET /items/?page=1&size=20&sort=price,asc&q=shoe → paged list

  - POST /items/ {name, price}

  - GET /items/{id}

  - PUT /items/{id} {name, price}

  - DELETE /items/{id}

ETag/304: GET /items/{id} returns ETag + Cache-Control; sends 304 Not Modified if If-None-Match matches.

OpenAPI: served at /openapi.yaml; Swagger UI at /docs.

Metrics: Prometheus at /metrics (optionally protected with basic auth via env).


## Security Notes
- Keep JWT secrets out of VCS; use .env only for local dev.
- Enable HTTPS/ingress in production.
- Restrict CORS (CORS_ORIGINS) to trusted origins.
- If exposing /metrics, protect with basic auth (env) or network policy.
- Redis is optional; when configured it stores refresh-token revocation and simple JSON caches.
- Rate-limit enabled by default; tune RATE_LIMIT_RPS/RATE_LIMIT_BURST.


## Common Issues
- Port already allocated: another app is using 8080 or 3001. Edit compose ports or stop the other app.
- Build context not found: set correct GOAPI_PATH / WEBAPP_PATH in infra/.env.
- DB connection errors: ensure postgres service is healthy; API depends_on healthcheck.
- 401 loops on frontend: verify VITE_API_URL and CORS; check X-Refresh-Token header on refresh.


## Common Issues
- Port already allocated: another app is using 8080 or 3001. Edit compose ports or stop the other app.
- Build context not found: set correct GOAPI_PATH / WEBAPP_PATH in infra/.env.
- DB connection errors: ensure postgres service is healthy; API depends_on healthcheck.
- 401 loops on frontend: verify VITE_API_URL and CORS; check X-Refresh-Token header on refresh.


















