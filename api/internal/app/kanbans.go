package app

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type kanbanColumn struct {
	ID            string    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Name          string    `json:"name"`
	Position      int       `json:"position"`
	ZendeskStatus string    `json:"zendesk_status"`
	Color         string    `json:"color"`
}

type kanbanBoard struct {
	ID        string         `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Name      string         `json:"name"`
	IsDefault bool           `json:"is_default"`
	Columns   []kanbanColumn `json:"columns"`
}

// @Summary     List Kanban boards
// @Description Returns all Kanban boards for the org, with columns nested, ordered by default first then name
// @Produce     json
// @Success     200  {array}   kanbanBoard
// @Failure     401  {string}  string  "Unauthorized"
// @Security    ApiKeyAuth
// @Router      /kanbans [get]
func (a *App) listKanbans(w http.ResponseWriter, r *http.Request) {
	o := orgFromContext(r.Context())

	boardRows, err := a.db.QueryContext(r.Context(), `
		SELECT id, created_at, updated_at, name, is_default
		FROM boards
		WHERE org_id = $1
		ORDER BY is_default DESC, name ASC
	`, o.ID)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("listKanbans boards query: %v", err)
		return
	}
	defer boardRows.Close()

	boards := []kanbanBoard{}
	// boardIndex tracks insertion order so we can attach columns efficiently
	boardIndex := map[string]int{}
	for boardRows.Next() {
		var b kanbanBoard
		if err := boardRows.Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt, &b.Name, &b.IsDefault); err != nil {
			http.Error(w, "scan failed", http.StatusInternalServerError)
			log.Printf("listKanbans boards scan: %v", err)
			return
		}
		b.Columns = []kanbanColumn{}
		boardIndex[b.ID] = len(boards)
		boards = append(boards, b)
	}

	if len(boards) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(boards)
		return
	}

	// Fetch all columns for this org's boards in one query
	colRows, err := a.db.QueryContext(r.Context(), `
		SELECT bc.id, bc.created_at, bc.updated_at, bc.board_id, bc.name, bc.position, bc.zendesk_status, bc.color
		FROM board_columns bc
		JOIN boards b ON b.id = bc.board_id
		WHERE b.org_id = $1
		ORDER BY bc.board_id, bc.position ASC
	`, o.ID)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("listKanbans columns query: %v", err)
		return
	}
	defer colRows.Close()

	for colRows.Next() {
		var c kanbanColumn
		var boardID string
		if err := colRows.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt, &boardID, &c.Name, &c.Position, &c.ZendeskStatus, &c.Color); err != nil {
			http.Error(w, "scan failed", http.StatusInternalServerError)
			log.Printf("listKanbans columns scan: %v", err)
			return
		}
		idx := boardIndex[boardID]
		boards[idx].Columns = append(boards[idx].Columns, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(boards)
}
