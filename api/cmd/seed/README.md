# seed

Wipes the database and inserts fake ticket data for development.

**Warning:** this deletes all existing data. Do not run against production.

## Usage

From the `api/` directory:

```
go run ./cmd/seed
```

Set `DATABASE_URL` to connect:

```
DATABASE_URL=postgres://user:pass@host:5433/db go run ./cmd/seed
```
