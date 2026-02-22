package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
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

type app struct {
	db    *sql.DB
	redis *redis.Client
	cfg   config
}

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

	a := &app{db: db, redis: rdb, cfg: cfg}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Get("/health", a.health)
	r.Options("/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r.Group(func(r chi.Router) {
		r.Use(a.requireAPIKey)
		r.Get("/org", a.getOrg)
		r.Get("/tickets", a.listTickets)
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-api-key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type contextKey string

const orgContextKey contextKey = "org"

type org struct {
	ID     string
	Name   string
	APIKey string
}

func orgFromContext(ctx context.Context) org {
	return ctx.Value(orgContextKey).(org)
}

func (a *app) requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("x-api-key")
		if key == "" {
			http.Error(w, "missing x-api-key header", http.StatusUnauthorized)
			return
		}

		var o org
		err := a.db.QueryRowContext(r.Context(), `
			SELECT id, name, api_key FROM organizations WHERE api_key = $1
		`, key).Scan(&o.ID, &o.Name, &o.APIKey)
		if err == sql.ErrNoRows {
			http.Error(w, "invalid api key", http.StatusUnauthorized)
			return
		}
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Printf("requireAPIKey query: %v", err)
			return
		}

		ctx := context.WithValue(r.Context(), orgContextKey, o)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type orgResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (a *app) getOrg(w http.ResponseWriter, r *http.Request) {
	o := orgFromContext(r.Context())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgResponse{ID: o.ID, Name: o.Name})
}

type ticketRow struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	Priority     string    `json:"priority"`
	ReporterName string    `json:"reporter_name"`
	CreatedAt    time.Time `json:"created_at"`
}

func (a *app) listTickets(w http.ResponseWriter, r *http.Request) {
	o := orgFromContext(r.Context())
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT t.id, t.title, t.description, t.status, t.priority, u.name, t.created_at
		FROM tickets t
		JOIN users u ON u.id = t.reporter_id
		WHERE t.org_id = $1
		ORDER BY t.created_at DESC
	`, o.ID)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("listTickets query: %v", err)
		return
	}
	defer rows.Close()

	tickets := []ticketRow{}
	for rows.Next() {
		var t ticketRow
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.ReporterName, &t.CreatedAt); err != nil {
			http.Error(w, "scan failed", http.StatusInternalServerError)
			log.Printf("listTickets scan: %v", err)
			return
		}
		tickets = append(tickets, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

func (a *app) health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dbOK := a.db.PingContext(ctx) == nil
	redisOK := a.redis.Ping(ctx).Err() == nil

	status := "ok"
	code := http.StatusOK
	if !dbOK || !redisOK {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{
		"status":   status,
		"postgres": dbOK,
		"redis":    redisOK,
	})
}
