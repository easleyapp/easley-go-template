# Easley dashboard starter (Go)

A small, portable Go application: an HTTP backend, a static frontend and a
SQLite database. It is the starting point for [Easley][1] application projects.

This repository is a **standalone Go module** with no dependency on the Easley
server..

## Run locally

```sh
cp .env.example .env
go run ./cmd/app
```

Then, open http://localhost:8080. The backend serves the frontend and `/api`. Environment values in
`.env` takes precedence.

## Contract

Easley builds and runs this project. Keep to this contract:

- Listen on `PORT` (default `8080`).
- Expose `GET /healthz`.
- Serve the frontend and API from the same server.
- Store durable data under `DATA_DIR` (Easley mounts `/data`).
- Read secrets from environment variables, never source files.
- Call external APIs from the backend, not the browser.
- Log to stdout/stderr.
- Run migrations explicitly at startup.

`easley.json` is the manifest file where template name/version, entrypoint,
required secrets, outbound hosts, and storage are defined.

## Layout

```text
cmd/app                 entrypoint, HTTP server, shutdown
internal/config         environment configuration
internal/dashboard      HTTP handlers, API, embedded frontend
internal/db             SQLite open + versioned migrations
internal/integrations   external API client
```

## Endpoints

- `GET /healthz` - readiness
- `GET /api/records` - stored records
- `POST /api/refresh` - fetch from the source and replace stored records

## Container

```sh
docker build -t dashboard-go .
docker run --rm -p 5000:5000 -v "$PWD/data:/data" dashboard-go
```

The image is a multi-stage build running a static binary as a non-root user. It
reads `PORT` (default `5000`), stores data under `/data`, and exposes `/healthz`.

## Test

```sh
go test ./...
```

## License

MIT. See [LICENSE](LICENSE).

[1]: https://easley.app
