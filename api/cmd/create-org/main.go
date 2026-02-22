package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

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

	var id string
	err = db.QueryRow(
		`INSERT INTO organizations (name, api_key) VALUES ($1, $2) RETURNING id`,
		name, apiKey,
	).Scan(&id)
	if err != nil {
		log.Fatalf("insert org: %v", err)
	}

	log.Printf("created org %q (id: %s)", name, id)
	log.Printf("api_key: %s", apiKey)
}
