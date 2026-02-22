# Purl

Support ticketing that doesn't make you want to close the tab.

Purl is a modern support ticketing system built for teams who are tired of Zendesk. Fast, beautiful, and actually pleasant to use.

## Features

- **Inbox** — unified ticket queue with smart prioritization
- **Kanban** — kanban-style view of the same tickets; Inbox and Kanban are two lenses on the same data, so a ticket always appears in both
- **Dashboard** — real-time metrics on volume, response times, and team performance
- **Reporting** — the charts your manager actually wants to see

## Getting Started

**1. Copy the API env file:**

```bash
cp api/.env.example api/.env
```

**2. Start Docker:**

```bash
docker compose up -d
```

This starts the API (port 9090), frontend (port 9091), PostgreSQL, and Redis.

**3. Seed the database** (wipes existing data and inserts fake tickets):

```bash
./cmd.sh seed
```

## Project Structure

This is an NPM workspaces monorepo.

| Workspace | Description |
|-----------|-------------|
| `purl/` | Main frontend app (Vue 3 + Vite) |
| `lib/` | Shared packages |

## Tech Stack

- [Vue 3](https://vuejs.org/) with `<script setup>` composition API
- [Vue Router](https://router.vuejs.org/)
- [Vite](https://vitejs.dev/)
- [lucide-vue-next](https://lucide.dev/) for icons
