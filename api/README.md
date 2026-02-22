# API

Go backend for Purl.

## Running

From the repo root, copy the env file and start all services:

```bash
cp api/.env.example api/.env
docker compose up -d
```

This starts the API on port `9090`, along with Postgres (port `5433`) and Redis (port `6380`). The API waits for both to be healthy before starting.

In development, [air](https://github.com/air-verse/air) watches for file changes and automatically regenerates docs and restarts the server.

## Health check

```bash
curl localhost:9090/health
```

Returns `200 OK` when both Postgres and Redis are reachable, `503` otherwise.

## API Docs

Swagger UI is served at `http://localhost:9090/docs/index.html` when the API is running.

The OpenAPI spec is generated from annotations in `main.go` using [swag](https://github.com/swaggo/swag). Docs are regenerated automatically on every build (air, Docker, CI) — no manual step required.

## Environment

See `.env.example` for all supported variables. The defaults work out of the box with `docker compose`.
