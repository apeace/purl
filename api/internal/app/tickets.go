package app

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type ticketRow struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ZendeskStatus *string   `json:"zendesk_status"`
	ReporterName  string    `json:"reporter_name"`
	CreatedAt     time.Time `json:"created_at"`
}

type ticketCommentRow struct {
	ID        string    `json:"id"`
	Body      string    `json:"body"`
	Channel   string    `json:"channel"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// @Summary     List tickets
// @Tags        Tickets
// @Description Returns all tickets for the org, ordered by creation date descending
// @Produce     json
// @Success     200  {array}   ticketRow
// @Failure     401  {string}  string  "Unauthorized"
// @Security    ApiKeyAuth
// @Router      /tickets [get]
func (a *App) listTickets(w http.ResponseWriter, r *http.Request) {
	o := orgFromContext(r.Context())
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT t.id, t.title, t.description, t.zendesk_status, c.name, t.created_at
		FROM tickets t
		JOIN customers c ON c.id = t.reporter_id
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
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.ZendeskStatus, &t.ReporterName, &t.CreatedAt); err != nil {
			http.Error(w, "scan failed", http.StatusInternalServerError)
			log.Printf("listTickets scan: %v", err)
			return
		}
		tickets = append(tickets, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

// @Summary     List ticket comments
// @Tags        Tickets
// @Description Returns all comments for a ticket, ordered by creation date ascending
// @Produce     json
// @Param       ticketID  path      string  true  "Ticket ID"
// @Success     200  {array}   ticketCommentRow
// @Failure     401  {string}  string  "Unauthorized"
// @Failure     404  {string}  string  "Not Found"
// @Security    ApiKeyAuth
// @Router      /tickets/{ticketID}/comments [get]
func (a *App) listTicketComments(w http.ResponseWriter, r *http.Request) {
	o := orgFromContext(r.Context())
	ticketID := chi.URLParam(r, "ticketID")

	// Verify the ticket belongs to this org
	var exists bool
	err := a.db.QueryRowContext(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1 AND org_id = $2)`,
		ticketID, o.ID,
	).Scan(&exists)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("listTicketComments check: %v", err)
		return
	}
	if !exists {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	rows, err := a.db.QueryContext(r.Context(), `
		SELECT id, body, channel::text, role::text, created_at
		FROM ticket_comments
		WHERE ticket_id = $1
		ORDER BY created_at ASC
	`, ticketID)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("listTicketComments query: %v", err)
		return
	}
	defer rows.Close()

	comments := []ticketCommentRow{}
	for rows.Next() {
		var c ticketCommentRow
		if err := rows.Scan(&c.ID, &c.Body, &c.Channel, &c.Role, &c.CreatedAt); err != nil {
			http.Error(w, "scan failed", http.StatusInternalServerError)
			log.Printf("listTicketComments scan: %v", err)
			return
		}
		comments = append(comments, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
