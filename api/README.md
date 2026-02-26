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

## Commands

All commands are run via `cmd.sh` at the repo root, which executes them inside the `api` Docker container with the correct environment. The format is:

```bash
./cmd.sh <command> [args...]
```

### create-org

Creates an org and generates its `api_key` and `zendesk_webhook_secret`. Usable on dev and prod.

```bash
./cmd.sh create-org "Acme Corp"
```

Prints the generated secrets — save them. The org slug is derived automatically from the name.

### reset-zendesk

Wipes all Zendesk-sourced data for an org (tickets, customers, agents, webhook events) and re-imports it fresh from the Zendesk API. Credentials are read from the database, so the org must already have them configured.

```bash
./cmd.sh reset-zendesk <slug>
```

### reset-orgs

Wipes and recreates orgs from `clients.json`, then imports Zendesk data for each. Use this locally to get a clean slate from a known config.

```bash
cp api/clients.example.json api/clients.json
# fill in credentials, then:
./cmd.sh reset-orgs
```

If `api_key` or `zendesk_webhook_secret` are empty in `clients.json`, random values are generated and printed — copy them into your Zendesk webhook config. Providing stable values in the JSON avoids having to reconfigure the webhook after each reset.

To point to a different file:

```bash
./cmd.sh reset-orgs -clients /path/to/clients.json
```

## Environment

See `.env.example` for all supported variables. The defaults work out of the box with `docker compose`.
