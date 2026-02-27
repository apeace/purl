package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatal("Usage: create-user <name> <email> <org-slug>")
	}

	name := os.Args[1]
	email := os.Args[2]
	orgSlug := os.Args[3]

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

	var orgID, orgName string
	err = db.QueryRow(`SELECT id, name FROM organizations WHERE slug = $1`, orgSlug).Scan(&orgID, &orgName)
	if err == sql.ErrNoRows {
		log.Fatalf("organization with slug %q not found", orgSlug)
	}
	if err != nil {
		log.Fatalf("look up org: %v", err)
	}

	var exists bool
	err = db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE org_id = $1 AND email = $2)`,
		orgID, email,
	).Scan(&exists)
	if err != nil {
		log.Fatalf("check user: %v", err)
	}
	if exists {
		log.Fatalf("user with email %q already exists in org %q", email, orgName)
	}

	var userID string
	err = db.QueryRow(
		`INSERT INTO users (name, email, org_id) VALUES ($1, $2, $3) RETURNING id`,
		name, email, orgID,
	).Scan(&userID)
	if err != nil {
		log.Fatalf("insert user: %v", err)
	}

	log.Printf("created user %q <%s> (id: %s) in org %q", name, email, userID, orgName)
}
