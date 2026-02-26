package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"purl/api/internal/app"
	"purl/api/internal/ratelimit"
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

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		log.Fatal("REDIS_URL environment variable is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("parse redis url: %v", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("ping redis: %v", err)
	}

	maxReqs := int64(100)
	if s := os.Getenv("ZENDESK_RATE_LIMIT"); s != "" {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			log.Fatalf("invalid ZENDESK_RATE_LIMIT: %v", err)
		}
		maxReqs = n
	}
	limiter := ratelimit.New(rdb, "zendesk", maxReqs, time.Minute)
	log.Printf("Zendesk rate limit: %d req/min", maxReqs)

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

	if err := app.ImportZendeskData(context.Background(), db, limiter, orgID, subdomain, email, apiKey); err != nil {
		log.Fatalf("import zendesk data: %v", err)
	}
}
