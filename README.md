# Full-Stack PostgreSQL: Go API + React + Docker

A minimal full-stack starter that runs entirely in Docker:
- **Backend:** Go (chi + pgx over `database/sql`)
- **DB:** PostgreSQL 17
- **Frontend:** React + Vite + TypeScript
- **Orchestration:** Docker Compose

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
go-api/
cmd/api/main.go
go.mod
internal/
config/config.go
db/db.go
domain/item.go
repo/item_repo.go
service/item_service.go
http/
cors.go
mw.go
handlers.go
router.go
Dockerfile
.env
webapp/
src/App.tsx
src/main.tsx
index.html
package.json
tsconfig.json
vite.config.ts
Dockerfile
.env
infra/
docker-compose.yml
postgres/
init/
01-create-db.sql
02-schema.sql
```

## Prerequisites

- Docker Desktop (Compose included)
- Node 20 LTS (only for local web dev)
- Go 1.23 (only if running backend outside Docker)

---

## Quick Start (Everything via Docker)

```bash
cd infra
docker compose up -d --build
# API:  http://localhost:8080/health  -> {"ok":true}
# Web:  http://localhost:3000
```