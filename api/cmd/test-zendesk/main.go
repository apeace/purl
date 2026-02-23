package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: test-zendesk <org-slug>")
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

	var subdomain, email, apiKey string
	err = db.QueryRow(
		`SELECT zendesk_subdomain, COALESCE(zendesk_email, ''), COALESCE(zendesk_api_key, '') FROM organizations WHERE slug = $1`,
		slug,
	).Scan(&subdomain, &email, &apiKey)
	if err == sql.ErrNoRows {
		log.Fatalf("no organization found with slug %q", slug)
	}
	if err != nil {
		log.Fatalf("query org: %v", err)
	}
	if subdomain == "" || email == "" || apiKey == "" {
		log.Fatalf("org %q has no Zendesk credentials configured", slug)
	}

	url := fmt.Sprintf("https://%s.zendesk.com/api/v2/users/me.json", subdomain)
	log.Printf("testing Zendesk API: GET %s", url)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("create request: %v", err)
	}
	// Zendesk API token auth requires Basic auth with "{email}/token:{api_token}" encoded in base64
	creds := base64.StdEncoding.EncodeToString([]byte(email + "/token:" + apiKey))
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("success: Zendesk API responded with %s", resp.Status)
	} else {
		log.Fatalf("Zendesk API returned %s — check credentials for org %q", resp.Status, slug)
	}
}
