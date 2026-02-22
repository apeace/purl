# seed

Wipes the database and inserts fake ticket data for development.

**Warning:** this deletes all existing data. Do not run against production.

## Usage

From the `api/` directory:

```
go run ./cmd/seed
```

By default it connects to `postgres://pipeline:pipeline@localhost:5432/pipeline`. Override with `DATABASE_URL`:

```
DATABASE_URL=postgres://user:pass@host:5432/db go run ./cmd/seed
```
