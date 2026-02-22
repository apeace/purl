# API — Claude Guidelines

## OpenAPI Docs

Handler annotations live in `main.go` as Go doc comments (swaggo/swag format).

The `api/docs/` directory is gitignored — docs are generated automatically at build time (air, Docker, CI). Never commit generated docs files.

Use `docker compose up -d` for local development. Air watches for changes and automatically reruns `swag init` and rebuilds the server.

When API annotations change, the frontend container's Vite plugin regenerates the TypeScript client automatically when the spec file changes. If the frontend is not running, restart it to pick up the new spec.
