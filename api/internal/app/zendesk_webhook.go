package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func (a *App) handleZendeskWebhook(w http.ResponseWriter, r *http.Request) {
	orgSlug := chi.URLParam(r, "orgSlug")

	// Look up org by slug, including the optional webhook secret.
	var orgID string
	var webhookSecret sql.NullString
	err := a.db.QueryRowContext(r.Context(), `
		SELECT id, zendesk_webhook_secret FROM organizations WHERE slug = $1
	`, orgSlug).Scan(&orgID, &webhookSecret)
	if err == sql.ErrNoRows {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("handleZendeskWebhook org lookup: %v", err)
		return
	}

	// Read the raw body so it can be used for signature verification and JSON decoding.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}

	// Verify HMAC-SHA256 signature when a secret is configured.
	if webhookSecret.Valid && webhookSecret.String != "" {
		timestamp := r.Header.Get("X-Zendesk-Webhook-Signature-Timestamp")
		signature := r.Header.Get("X-Zendesk-Webhook-Signature")

		mac := hmac.New(sha256.New, []byte(webhookSecret.String))
		mac.Write([]byte(timestamp))
		mac.Write(body)
		expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(expected), []byte(signature)) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
	} else {
		log.Printf("handleZendeskWebhook: no webhook secret configured for org %s — skipping signature verification", orgSlug)
	}

	var payload ZendeskWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = a.db.ExecContext(r.Context(), `
		INSERT INTO zendesk_webhook_events (org_id, event_id, event_type, payload)
		VALUES ($1, $2, $3, $4)
	`, orgID, payload.ID, payload.Type, body)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("handleZendeskWebhook: insert event: %v", err)
		return
	}

	a.dispatchZendeskEvent(r, orgID, &payload)

	w.WriteHeader(http.StatusOK)
}

func (a *App) dispatchZendeskEvent(r *http.Request, orgID string, payload *ZendeskWebhookPayload) {
	switch {
	case payload.Type == EventTypeUserCreated:
		a.handleZendeskUserCreated(r, orgID, payload)

	case strings.HasPrefix(payload.Type, "zen:event-type:ticket."):
		var detail ZendeskTicketDetail
		if err := json.Unmarshal(payload.Detail, &detail); err != nil {
			log.Printf("handleZendeskWebhook: failed to parse ticket detail for %s: %v", payload.Type, err)
			return
		}
		log.Printf("zendesk ticket event %s: ticket %s", payload.Type, detail.ID)

	case strings.HasPrefix(payload.Type, "zen:event-type:user."):
		log.Printf("zendesk user event not handled: type=%s id=%s", payload.Type, payload.ID)

	default:
		log.Printf("zendesk event not handled: type=%s id=%s", payload.Type, payload.ID)
	}
}

func (a *App) handleZendeskUserCreated(r *http.Request, orgID string, payload *ZendeskWebhookPayload) {
	var detail ZendeskUserDetail
	if err := json.Unmarshal(payload.Detail, &detail); err != nil {
		log.Printf("handleZendeskUserCreated: failed to parse user detail: %v", err)
		return
	}

	switch detail.Role {
	case "end-user":
		// Insert a new customer and their email.
		var customerID string
		err := a.db.QueryRowContext(r.Context(), `
			INSERT INTO customers (name, org_id) VALUES ($1, $2) RETURNING id
		`, detail.Name, orgID).Scan(&customerID)
		if err != nil {
			log.Printf("handleZendeskUserCreated: insert customer: %v", err)
			return
		}
		_, err = a.db.ExecContext(r.Context(), `
			INSERT INTO customer_emails (customer_id, email, verified) VALUES ($1, $2, false)
		`, customerID, detail.Email)
		if err != nil {
			log.Printf("handleZendeskUserCreated: insert customer_email: %v", err)
		}

	case "agent", "admin":
		_, err := a.db.ExecContext(r.Context(), `
			INSERT INTO agents (email, name, org_id) VALUES ($1, $2, $3)
			ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name
		`, detail.Email, detail.Name, orgID)
		if err != nil {
			log.Printf("handleZendeskUserCreated: upsert agent: %v", err)
		}

	default:
		log.Printf("handleZendeskUserCreated: unknown role %q for user %s", detail.Role, detail.ID)
	}
}
