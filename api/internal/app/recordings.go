package app

import (
	"database/sql"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// proxyRecording streams a Zendesk call recording to the client.
// Uses query-param auth (?api_key=) so <audio> elements can reference it directly.
// The raw recording_url is never exposed — the frontend hits this proxy instead.
func (a *App) proxyRecording(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("api_key")
	if key == "" {
		http.Error(w, "missing api_key", http.StatusUnauthorized)
		return
	}

	var o org
	err := a.db.QueryRowContext(r.Context(),
		`SELECT id, name, api_key FROM organizations WHERE api_key = $1`, key,
	).Scan(&o.ID, &o.Name, &o.APIKey)
	if err == sql.ErrNoRows {
		http.Error(w, "invalid api key", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("proxyRecording auth: %v", err)
		return
	}

	ticketID := chi.URLParam(r, "ticketID")
	commentID := chi.URLParam(r, "commentID")

	// Look up recording URL, verifying the comment belongs to this org's ticket
	var recordingURL string
	err = a.db.QueryRowContext(r.Context(), `
		SELECT tc.recording_url
		FROM ticket_comments tc
		JOIN tickets t ON t.id = tc.ticket_id
		WHERE tc.id = $1 AND tc.ticket_id = $2 AND t.org_id = $3
		  AND tc.recording_url IS NOT NULL
	`, commentID, ticketID, o.ID).Scan(&recordingURL)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Fetch Zendesk credentials for this org
	var subdomain, email, apiKey string
	err = a.db.QueryRowContext(r.Context(),
		`SELECT COALESCE(zendesk_subdomain,''), COALESCE(zendesk_email,''), COALESCE(zendesk_api_key,'')
		 FROM organizations WHERE id = $1`,
		o.ID,
	).Scan(&subdomain, &email, &apiKey)
	if err != nil || subdomain == "" || email == "" || apiKey == "" {
		http.Error(w, "zendesk not configured", http.StatusInternalServerError)
		if err != nil {
			log.Printf("proxyRecording creds: %v", err)
		}
		return
	}

	// Stream from Zendesk
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, recordingURL, nil)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		log.Printf("proxyRecording newreq: %v", err)
		return
	}
	req.SetBasicAuth(email+"/token", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "upstream error", http.StatusBadGateway)
		log.Printf("proxyRecording fetch: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "upstream error", http.StatusBadGateway)
		log.Printf("proxyRecording upstream status: %d", resp.StatusCode)
		return
	}

	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	w.Header().Set("Cache-Control", "private, max-age=3600")
	io.Copy(w, resp.Body)
}
