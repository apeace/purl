package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	_ "purl/api/docs"
	"purl/api/internal/app"
)

//go:embed migrations/*.sql
var migrations embed.FS

type config struct {
	DatabaseURL string
	RedisURL    string
	Port        string
}

func loadConfig() config {
	return config{
		DatabaseURL: requireEnv("DATABASE_URL"),
		RedisURL:    requireEnv("REDIS_URL"),
		Port:        getEnv("PORT", "9090"),
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s environment variable is required", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

// @title           Purl API
// @version         1.0
// @description     API for Purl.
// @host            localhost:9090
// @securityDefinitions.apikey ApiKeyAuth
// @in              header
// @name            x-api-key
func main() {
	cfg := loadConfig()

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("connected to postgres")

	goose.SetBaseFS(migrations)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose set dialect: %v", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("goose up: %v", err)
	}
	log.Println("migrations applied")

	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("parse redis url: %v", err)
	}
	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("ping redis: %v", err)
	}
	log.Println("connected to redis")

	a := app.New(db, rdb)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, a.Handler()); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
