# go-vite-template

A full-stack app template: **Go backend + React/Vite frontend + SQLite**, built for fast iteration and single-binary deployment.

## What you get

- **`make dev`** — Go backend with hot reload (air) + Vite frontend with HMR, running in parallel
- **`make build`** — Single binary with the frontend embedded via `go:embed`
- **SQLite** with a numbered migration system — no external database
- **Bearer token auth** — admin token auto-generated on first run
- **GitHub Actions** — CI (typecheck + test + smoke), DCO sign-off check, tag-triggered release (linux + darwin binaries)

## Quick start

### Prerequisites

- Go 1.24+
- Node.js 20+
- [air](https://github.com/air-verse/air) (`go install github.com/air-verse/air@latest`)

### Development

```bash
cp .env.example .env   # optional — defaults work fine
npm install
make dev               # starts backend (:9002) + frontend (:9005)
```

Open http://localhost:9002. On first run, get your admin token:

```bash
# In another terminal:
make build
./dist/go-vite-template admin-token --raw
```

### Production

```bash
make build
./dist/go-vite-template --port 9002
```

### Install as a service

```bash
make install           # builds + installs binary + launchd/systemd service
make service-start
make service-logs
```

## Project structure

```
├── server-go/            # Go backend
│   ├── main.go           # HTTP server, router, embed, graceful shutdown
│   ├── config/           # Env config + admin auth
│   ├── db/               # SQLite + migrations
│   ├── middleware/        # Logging, CORS, bearer auth
│   └── modules/          # Feature modules (health, tokens, ...)
├── client/               # React + Vite frontend
│   └── src/
├── shared/               # Shared types (Zod schemas)
├── Makefile              # dev, build, run, test, install, service-*
└── .github/workflows/    # CI, DCO, Release
```

## LAN access / reverse proxy

To expose the dev server on a LAN address or run behind a reverse proxy (Caddy, nginx, etc.):

```bash
# .env
HOST=192.168.1.5   # or 0.0.0.0 for all interfaces
PORT=9002
```

Both the Go backend and Vite dev server will bind to `HOST`. Vite is configured with `allowedHosts: true`, so requests arriving via a proxy hostname won't be rejected.

For production behind a reverse proxy, pass the flags directly:

```bash
./dist/go-vite-template --port 9002 --host 0.0.0.0
```

## Customizing

1. Change `APP_NAME` in the Makefile
2. Change `AppName` in `server-go/config/config.go`
3. Change `module` in `server-go/go.mod`
4. Add your domain modules under `server-go/modules/`
5. Add migrations in `server-go/db/db.go`
6. Build your frontend in `client/src/`
