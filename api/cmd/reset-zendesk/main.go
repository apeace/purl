package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"purl/api/internal/app"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: reset-zendesk <org-slug>")
	}

	slug := os.Args[1]

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

	var orgID, subdomain, email, apiKey string
	err = db.QueryRow(
		`SELECT id, zendesk_subdomain, COALESCE(zendesk_email, ''), COALESCE(zendesk_api_key, '') FROM organizations WHERE slug = $1`,
		slug,
	).Scan(&orgID, &subdomain, &email, &apiKey)
	if err == sql.ErrNoRows {
		log.Fatalf("no organization found with slug %q", slug)
	}
	if err != nil {
		log.Fatalf("query org: %v", err)
	}
	if subdomain == "" || email == "" || apiKey == "" {
		log.Fatalf("org %q has no Zendesk credentials configured", slug)
	}

	if err := app.ImportZendeskData(context.Background(), db, orgID, subdomain, email, apiKey); err != nil {
		log.Fatalf("import zendesk data: %v", err)
	}
}
