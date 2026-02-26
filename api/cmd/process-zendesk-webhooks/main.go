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

	n, err := app.ProcessPendingWebhooks(context.Background(), db, limiter)
	if err != nil {
		log.Fatalf("process webhooks: %v", err)
	}
	if n > 0 {
		log.Printf("processed %d webhook event(s)", n)
	}
}
