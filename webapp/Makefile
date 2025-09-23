.PHONY: up down logs rebuild api web

up:    ; cd infra && docker compose up -d --build
logs:  ; cd infra && docker compose logs -f
down:  ; cd infra && docker compose down -v
api:   ; cd go-api && PORT=8080 DB_HOST=127.0.0.1 go run ./cmd/api
web:   ; cd webapp && npm run dev
