# AGENTS.md - Easley dashboard starter

This is a normal Go project. Use the filesystem, shell and tests directly;
Easley only owns deployment and runtime. Do not restructure the deployment
contract in `easley.json`.

## Rules

- Keep the app small. Prefer the standard library and a few pinned dependencies.
- Serve the frontend and `/api` from the same process.
- Bind to `PORT`; expose `GET /healthz`.
- Persist durable data under `DATA_DIR` (production `/data`) in SQLite.
- Read secrets from environment variables. Never write them into source or logs.
- Call external services from the backend. The browser only talks to this server.
- Only add outbound hosts that are declared in `easley.json`.
- Add schema changes as new files in `internal/db/migrations/`; never edit an
  applied migration.
- Run `gofmt`, `go build ./...`, `go vet ./...` and `go test ./...` before deploying.

## Commands

```sh
go run ./cmd/app
go build ./...
go vet ./...
go test ./...
```
