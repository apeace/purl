# API

Go backend for Purl.

## Running

### With Docker (recommended)

From the repo root, copy the env file and start all services:

```bash
cp api/.env.example api/.env
docker-compose up --build
```

This starts the API on port `9090`, along with Postgres (port `5433`) and Redis (port `6380`). The API waits for both to be healthy before starting.

### Locally

If you want to run the API process directly (e.g. for faster iteration), start just the infra via Docker and run the API with Go:

```bash
# From repo root — start Postgres and Redis only
docker-compose up postgres redis

# From api/ — run the API
cp .env.example .env
go run .
```

## Health check

```bash
curl localhost:9090/health
```

Returns `200 OK` when both Postgres and Redis are reachable, `503` otherwise.

## Environment

See `.env.example` for all supported variables. The defaults work out of the box with `docker-compose`.
