# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A full-stack app template with a React/TypeScript frontend and Go backend. Single binary deployment — the frontend is embedded into the Go binary via `go:embed`. SQLite for storage, bearer-token auth, auto-generated admin token on first run.

## Commands

- `make dev` — Start dev servers (Go backend on :9002 + Vite frontend on :9005 with proxy)
- `make build` — Build frontend (Vite → dist/public/) and backend (Go → dist/__APP_NAME__)
- `make run` — Run production build
- `npm run check` — TypeScript type checking (frontend only)

### Individual dev servers
- `make dev-backend` — Go backend only (with DEV=1 for Vite proxy mode, requires `air`)
- `make dev-frontend` — Vite dev server only (proxies /api to backend)

## Architecture

### Go Backend (`server-go/`)

Modular Go backend using chi router and SQLite. Bearer-token auth via auto-generated admin token.

```
server-go/
  main.go              # HTTP server, middleware, Vite proxy (dev) / static files (prod)
  config/config.go     # Environment config (PORT, DATA_DIR, DB_PATH, DEV), .env loading
  config/adminauth.go  # Auto-generated admin token (DATA_DIR/adminAuth.json)
  db/db.go             # SQLite connection + numbered schema migrations
  middleware/logging.go # Request logging + CORS
  middleware/auth.go    # Bearer token validation
  modules/
    health/            # Health check endpoint (example module)
    tokens/            # API token CRUD
```

Each module follows: `model.go` (types), `storage.go` (DB queries), `handler.go` (HTTP handlers).

### API

All endpoints are versioned under `/api/v1/` and require a Bearer token.

### Shared Layer
- `shared/schema.ts` — Source of truth for Zod validation schemas used by the frontend. DB schema is owned by Go (`db/db.go`).

### Frontend (`client/src/`)
- `App.tsx` — Router: token entry gate or main app
- `lib/queryClient.ts` — API helper with auth header, TanStack Query client
- UI: React + Tailwind CSS

### Path Aliases
- `@/*` → `client/src/`
- `@shared/*` → `shared/`

### Environment Variables
- `PORT` — Server port (default: 9002)
- `HOST` — Comma-separated bind addresses (default: `127.0.0.1`)
- `DATA_DIR` — Base directory for DB and files (default: `$XDG_DATA_HOME/__APP_NAME__`)
- `DB_PATH` — SQLite database path (default: `DATA_DIR/__APP_NAME__.db`)
- `DEV` — Enable Vite dev proxy mode
- `VITE_URL` — Vite dev server URL (default: http://127.0.0.1:9005)
- `AGENT_CMD` — AI agent command for the web terminal (e.g., `claude`)
- `WORK_DIR` — Working directory for the agent process (default: current directory)

## Design Guidelines

- JetBrains Mono for all text
- HSL CSS variables for theming (dark mode via `.dark` class)
- Minimalist aesthetic, generous whitespace
