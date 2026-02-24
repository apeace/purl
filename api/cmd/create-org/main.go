package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

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

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: create-org <name>")
	}

	name := os.Args[1]

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

	var exists bool
	err = db.QueryRow(`SELECT EXISTS(SELECT 1 FROM organizations WHERE name = $1)`, name).Scan(&exists)
	if err != nil {
		log.Fatalf("check org: %v", err)
	}
	if exists {
		log.Fatalf("organization %q already exists", name)
	}

	// 32 random bytes → 64-character hex string
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		log.Fatalf("generate api key: %v", err)
	}
	apiKey := hex.EncodeToString(keyBytes)

	secretBytes := make([]byte, 32)
	if _, err := rand.Read(secretBytes); err != nil {
		log.Fatalf("generate webhook secret: %v", err)
	}
	webhookSecret := hex.EncodeToString(secretBytes)

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}
	defer tx.Rollback()

	var orgID string
	err = tx.QueryRow(
		`INSERT INTO organizations (name, api_key, zendesk_webhook_secret) VALUES ($1, $2, $3) RETURNING id`,
		name, apiKey, webhookSecret,
	).Scan(&orgID)
	if err != nil {
		log.Fatalf("insert org: %v", err)
	}

	var boardID string
	err = tx.QueryRow(
		`INSERT INTO boards (org_id, name, is_default) VALUES ($1, $2, true) RETURNING id`,
		orgID, "Ticket Status",
	).Scan(&boardID)
	if err != nil {
		log.Fatalf("insert default board: %v", err)
	}

	for _, col := range defaultBoardColumns {
		_, err = tx.Exec(
			`INSERT INTO board_columns (board_id, name, position, zendesk_status, color)
			 VALUES ($1, $2, $3, $4, $5)`,
			boardID, col.name, col.position, col.zendeskStatus, col.color,
		)
		if err != nil {
			log.Fatalf("insert board column %q: %v", col.name, err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("commit tx: %v", err)
	}

	log.Printf("created org %q (id: %s)", name, orgID)
	log.Printf("api_key: %s", apiKey)
	log.Printf("zendesk_webhook_secret: %s", webhookSecret)
}
