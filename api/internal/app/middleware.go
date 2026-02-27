package app

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, x-api-key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type contextKey string

const orgContextKey contextKey = "org"
const userContextKey contextKey = "user"
const sessionTokenContextKey contextKey = "sessionToken"

type org struct {
	ID               string
	Name             string
	APIKey           string
	ZendeskSubdomain string
}

type user struct {
	ID      string
	Name    string
	Email   string
	OrgID   string
	OrgName string
}

func orgFromContext(ctx context.Context) org {
	return ctx.Value(orgContextKey).(org)
}

func userFromContext(ctx context.Context) user {
	return ctx.Value(userContextKey).(user)
}

func sessionTokenFromContext(ctx context.Context) string {
	v, _ := ctx.Value(sessionTokenContextKey).(string)
	return v
}

func (a *App) requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("x-api-key")
		if key == "" {
			http.Error(w, "missing x-api-key header", http.StatusUnauthorized)
			return
		}

		var o org
		err := a.db.QueryRowContext(r.Context(), `
			SELECT id, name, api_key, COALESCE(zendesk_subdomain, '') FROM organizations WHERE api_key = $1
		`, key).Scan(&o.ID, &o.Name, &o.APIKey, &o.ZendeskSubdomain)
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

// requireSession validates a Bearer token from the Authorization header against Redis sessions.
func (a *App) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			http.Error(w, "missing authorization", http.StatusUnauthorized)
			return
		}

		u, err := a.loadSessionUser(r, token)
		if err != nil {
			http.Error(w, "invalid or expired session", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, userContextKey, u)
		ctx = context.WithValue(ctx, sessionTokenContextKey, token)
		ctx = context.WithValue(ctx, orgContextKey, org{ID: u.OrgID, Name: u.OrgName})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireAuth tries Bearer token (session) first, then falls back to x-api-key.
func (a *App) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try session auth first
		if token := extractBearerToken(r); token != "" {
			u, err := a.loadSessionUser(r, token)
			if err == nil {
				ctx := r.Context()
				ctx = context.WithValue(ctx, userContextKey, u)
				ctx = context.WithValue(ctx, sessionTokenContextKey, token)
				ctx = context.WithValue(ctx, orgContextKey, org{ID: u.OrgID, Name: u.OrgName})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Fall back to API key
		key := r.Header.Get("x-api-key")
		if key == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var o org
		err := a.db.QueryRowContext(r.Context(), `
			SELECT id, name, api_key, COALESCE(zendesk_subdomain, '') FROM organizations WHERE api_key = $1
		`, key).Scan(&o.ID, &o.Name, &o.APIKey, &o.ZendeskSubdomain)
		if err == sql.ErrNoRows {
			http.Error(w, "invalid api key", http.StatusUnauthorized)
			return
		}
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Printf("requireAuth apikey query: %v", err)
			return
		}

		ctx := context.WithValue(r.Context(), orgContextKey, o)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return ""
}

func (a *App) loadSessionUser(r *http.Request, token string) (user, error) {
	userID, err := a.redis.Get(r.Context(), "session:"+token).Result()
	if err != nil {
		return user{}, err
	}

	var u user
	err = a.db.QueryRowContext(r.Context(), `
		SELECT u.id, u.name, u.email, u.org_id, o.name
		FROM users u
		JOIN organizations o ON o.id = u.org_id
		WHERE u.id = $1
	`, userID).Scan(&u.ID, &u.Name, &u.Email, &u.OrgID, &u.OrgName)
	if err != nil {
		return user{}, err
	}

	return u, nil
}
