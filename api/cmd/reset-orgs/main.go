package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	"purl/api/internal/app"
	"purl/api/internal/ratelimit"
)

type client struct {
	Slug                 string `json:"slug"`
	Name                 string `json:"name"`
	APIKey               string `json:"api_key"`
	ZendeskWebhookSecret string `json:"zendesk_webhook_secret"`
	ZendeskSubdomain     string `json:"zendesk_subdomain"`
	ZendeskEmail         string `json:"zendesk_email"`
	ZendeskAPIKey        string `json:"zendesk_api_key"`
}

type boardColumn struct {
	name          string
	position      int
	zendeskStatus string
	color         string
}

var defaultBoardColumns = []boardColumn{
	{name: "New", position: 0, zendeskStatus: "new", color: "#1F73B7"},
	{name: "Open", position: 1, zendeskStatus: "open", color: "#CC3340"},
	{name: "Pending", position: 2, zendeskStatus: "pending", color: "#AD5E18"},
	{name: "Solved", position: 3, zendeskStatus: "solved", color: "#228F67"},
	{name: "Closed", position: 4, zendeskStatus: "closed", color: "#68737D"},
}

func generateHex() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("generate random: %v", err)
	}
	return hex.EncodeToString(b)
}

func resetOrg(db *sql.DB, limiter *ratelimit.Limiter, c client) {
	// Wipe all org data in FK-safe order. Using a subquery for org_id means
	// each statement is a no-op if the org doesn't exist yet.
	// Cascades: tickets→ticket_comments,board_tickets; customers→customer_emails,customer_phones; boards→board_columns
	wipes := []string{
		`DELETE FROM zendesk_webhook_events WHERE org_id = (SELECT id FROM organizations WHERE slug = $1)`,
		`DELETE FROM tickets       WHERE org_id = (SELECT id FROM organizations WHERE slug = $1)`,
		`DELETE FROM customers     WHERE org_id = (SELECT id FROM organizations WHERE slug = $1)`,
		`DELETE FROM agents        WHERE org_id = (SELECT id FROM organizations WHERE slug = $1)`,
		`DELETE FROM boards        WHERE org_id = (SELECT id FROM organizations WHERE slug = $1)`,
		`DELETE FROM organizations WHERE slug = $1`,
	}
	for _, stmt := range wipes {
		if _, err := db.Exec(stmt, c.Slug); err != nil {
			log.Fatalf("[%s] wipe: %v", c.Slug, err)
		}
	}

	apiKey := c.APIKey
	if apiKey == "" {
		apiKey = generateHex()
		log.Printf("[%s] generated api_key: %s", c.Slug, apiKey)
	}
	webhookSecret := c.ZendeskWebhookSecret
	if webhookSecret == "" {
		webhookSecret = generateHex()
		log.Printf("[%s] generated zendesk_webhook_secret: %s", c.Slug, webhookSecret)
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("[%s] begin tx: %v", c.Slug, err)
	}
	defer tx.Rollback()

	var orgID string
	err = tx.QueryRow(
		`INSERT INTO organizations (name, api_key, zendesk_webhook_secret, zendesk_subdomain, zendesk_email, zendesk_api_key)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id`,
		c.Name, apiKey, webhookSecret, c.ZendeskSubdomain, c.ZendeskEmail, c.ZendeskAPIKey,
	).Scan(&orgID)
	if err != nil {
		log.Fatalf("[%s] insert org: %v", c.Slug, err)
	}

	var boardID string
	err = tx.QueryRow(
		`INSERT INTO boards (org_id, name, is_default) VALUES ($1, $2, true) RETURNING id`,
		orgID, "Ticket Status",
	).Scan(&boardID)
	if err != nil {
		log.Fatalf("[%s] insert default board: %v", c.Slug, err)
	}

	for _, col := range defaultBoardColumns {
		_, err = tx.Exec(
			`INSERT INTO board_columns (board_id, name, position, zendesk_status, color)
			 VALUES ($1, $2, $3, $4, $5)`,
			boardID, col.name, col.position, col.zendeskStatus, col.color,
		)
		if err != nil {
			log.Fatalf("[%s] insert board column %q: %v", c.Slug, col.name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("[%s] commit tx: %v", c.Slug, err)
	}
	log.Printf("[%s] created org (id: %s)", c.Slug, orgID)

	if err := app.ImportZendeskData(context.Background(), db, limiter, orgID, c.ZendeskSubdomain, c.ZendeskEmail, c.ZendeskAPIKey); err != nil {
		log.Fatalf("[%s] import zendesk data: %v", c.Slug, err)
	}
}

func main() {
	// PURL_CLIENTS_JSON is set by cmd.sh to pass the file contents across the
	// Docker boundary without relying on volume mounts.
	jsonStr := os.Getenv("PURL_CLIENTS_JSON")
	if jsonStr == "" {
		log.Fatal("PURL_CLIENTS_JSON is not set; run via ./cmd.sh")
	}
	var clients []client
	if err := json.Unmarshal([]byte(jsonStr), &clients); err != nil {
		log.Fatalf("parse clients: %v", err)
	}
	if len(clients) == 0 {
		log.Fatal("no clients found in PURL_CLIENTS_JSON")
	}

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

	for _, c := range clients {
		log.Printf("resetting org %q...", c.Slug)
		resetOrg(db, limiter, c)
	}
	log.Printf("done — reset %d org(s)", len(clients))
}
