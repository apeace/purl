package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"purl/api/internal/app"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	n, err := app.ProcessPendingWebhooks(context.Background(), db)
	if err != nil {
		log.Fatalf("process webhooks: %v", err)
	}
	if n > 0 {
		log.Printf("processed %d webhook event(s)", n)
	}
}
