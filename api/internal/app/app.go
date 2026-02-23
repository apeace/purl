package app

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
)

// App holds shared dependencies for all handlers.
type App struct {
	db    *sql.DB
	redis *redis.Client
}

// New constructs an App with the given database and Redis client.
func New(db *sql.DB, rdb *redis.Client) *App {
	return &App{db: db, redis: rdb}
}

// Handler builds and returns the chi router with all middleware and routes registered.
func (a *App) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusMovedPermanently)
	})
	r.Get("/docs/*", httpSwagger.Handler())
	r.Get("/health", a.health)
	r.Options("/*", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r.Post("/webhooks/zendesk/{orgSlug}", a.handleZendeskWebhook)

	r.Group(func(r chi.Router) {
		r.Use(a.requireAPIKey)
		r.Get("/org", a.getOrg)
		r.Get("/tickets", a.listTickets)
	})

	return r
}
