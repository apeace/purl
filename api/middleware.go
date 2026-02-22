package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
)

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
