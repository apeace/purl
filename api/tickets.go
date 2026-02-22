package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type ticketRow struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	Priority     string    `json:"priority"`
	ReporterName string    `json:"reporter_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// @Summary     List tickets
// @Description Returns all tickets for the org, ordered by creation date descending
// @Produce     json
// @Success     200  {array}   ticketRow
// @Failure     401  {string}  string  "Unauthorized"
// @Security    ApiKeyAuth
// @Router      /tickets [get]
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
