# API

Go backend for Pipeline.

## Running

### With Docker (recommended)

From the repo root, copy the env file and start all services:

```bash
cp api/.env.example api/.env
docker-compose up --build
```

This starts the API on port `8080`, along with Postgres (port `5432`) and Redis (port `6379`). The API waits for both to be healthy before starting.

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
curl localhost:8080/health
```

Returns `200 OK` when both Postgres and Redis are reachable, `503` otherwise.

## Environment

See `.env.example` for all supported variables. The defaults work out of the box with `docker-compose`.
