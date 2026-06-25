# template-v3

A full-stack starter template for building web apps with a Go backend and SvelteKit frontend.

**Backend** — [huma v2](https://github.com/danielgtaylor/huma) (OpenAPI-first) · chi router · sqlc · pgx (PostgreSQL)  
**Frontend** — SvelteKit SPA · TanStack Query · shadcn-svelte · TypeScript client auto-generated from the OpenAPI spec

## What's included

- **Email OTP auth** — passwordless login with session tokens; `123456` works on localhost
- **API keys** — create and manage API keys per user
- **Companies** — multi-tenant company model with member management
- **Feature flags** — per-user flag reads with admin upsert
- **OpenAPI spec** — committed at `openapi.json`, regenerated on every `make gen`
- **TypeScript SDK** — auto-generated from the spec, no manual API client code
- **Docker** — single-stage multi-arch Dockerfile builds a ~20MB binary

## Prerequisites

- Go 1.22+
- [Bun](https://bun.sh)
- [sqlc](https://sqlc.dev) — `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- PostgreSQL (or Docker)

## Getting started

```bash
# 1. Copy env file and fill in your DATABASE_URL and SMTP settings
cp .env.example .env

# 2. Run the server (auto-migrates on startup)
make run          # Go API on :8080

# 3. In a separate terminal, start the frontend dev server
cd frontend
bun install
bun run dev       # Vite on :5173, proxies /api → :8080
```

Open `http://localhost:5173`. Enter any email — check your terminal for the OTP (or use `123456`).

## Key commands

| Command | What it does |
|---------|-------------|
| `make run` | Start the Go server on `:8080` |
| `make gen` | Regenerate sqlc queries + `openapi.json` + TypeScript SDK |
| `make build` | Full build: gen → frontend → `bin/server` binary |
| `make check` | `go vet` + JSON tag lint + `tsc` |
| `make migrate name=add_foo` | Create a new up/down migration pair |
| `cd frontend && bun run dev` | Vite dev server on `:5173` |
| `cd frontend && bun run check` | Svelte type-check (target: 0 errors) |

## Project structure

```
cmd/server/          — main binary
internal/
  auth/              — OTP login, session tokens
  apikey/            — API key CRUD
  user/              — /api/users/me
  company/           — company + member management
  featureflag/       — feature flag reads and admin upsert
  platform/
    compiled/        — sqlc-generated queries (do not edit)
    migration/sql/   — SQL migration files
    humax/           — auth helpers, error helpers
frontend/src/
  lib/api/           — generated SDK (do not edit)
  lib/api-client.ts  — SDK config (base URL, auth header)
  routes/app/        — authenticated pages
  routes/(auth)/     — login / verify-otp pages
```

## Adding a feature

Use the `/build` skill in Claude Code — it takes a brief, asks clarifying questions, and implements the full backend + frontend in the correct sequence for this repo.

## Deploying

```bash
docker build -t my-app .
docker run -e DATABASE_URL=... -e SMTP_HOST=... -p 8080:8080 my-app
```

The container serves both the API and the built frontend from a single binary on port `8080`.
