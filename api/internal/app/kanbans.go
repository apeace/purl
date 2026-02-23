package app

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
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

type createKanbanRequest struct {
	Name string `json:"name"`
}

// @Summary     Create a Kanban board
// @Description Creates a new Kanban board for the org
// @Accept      json
// @Produce     json
// @Param       body  body      createKanbanRequest  true  "Board to create"
// @Success     201   {object}  kanbanBoard
// @Failure     400   {string}  string  "Bad Request"
// @Failure     401   {string}  string  "Unauthorized"
// @Security    ApiKeyAuth
// @Router      /kanbans [post]
func (a *App) createKanban(w http.ResponseWriter, r *http.Request) {
	o := orgFromContext(r.Context())

	var req createKanbanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	var b kanbanBoard
	err := a.db.QueryRowContext(r.Context(), `
		INSERT INTO boards (org_id, name, is_default)
		VALUES ($1, $2, false)
		RETURNING id, created_at, updated_at, name, is_default
	`, o.ID, req.Name).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt, &b.Name, &b.IsDefault)
	if err != nil {
		http.Error(w, "insert failed", http.StatusInternalServerError)
		log.Printf("createKanban insert: %v", err)
		return
	}
	b.Columns = []kanbanColumn{}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(b)
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

// requireBoardInOrg checks that the board URL param exists and belongs to the org.
// Returns the board ID on success, or writes a 404 and returns "".
func (a *App) requireBoardInOrg(w http.ResponseWriter, r *http.Request) string {
	boardID := chi.URLParam(r, "boardID")
	o := orgFromContext(r.Context())

	var exists bool
	err := a.db.QueryRowContext(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM boards WHERE id = $1 AND org_id = $2)`,
		boardID, o.ID,
	).Scan(&exists)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("requireBoardInOrg query: %v", err)
		return ""
	}
	if !exists {
		http.Error(w, "board not found", http.StatusNotFound)
		return ""
	}
	return boardID
}

// putColumnItem is one element of a PUT /kanbans/{boardID}/columns request.
// ID is nil for new columns and non-nil for existing ones being updated.
type putColumnItem struct {
	ID            *string `json:"id"`
	Name          string  `json:"name"`
	Position      int     `json:"position"`
	ZendeskStatus string  `json:"zendesk_status"`
	Color         string  `json:"color"`
}

// @Summary     Replace board columns
// @Description Atomically replaces all columns on a Kanban board.
// @Description Columns with an id are updated; columns without an id are created;
// @Description columns present in the database but absent from the request are deleted.
// @Accept      json
// @Produce     json
// @Param       boardID  path      string           true  "Board ID"
// @Param       body     body      []putColumnItem  true  "Full desired column list"
// @Success     200      {array}   kanbanColumn
// @Failure     400      {string}  string  "Bad Request"
// @Failure     401      {string}  string  "Unauthorized"
// @Failure     404      {string}  string  "Not Found"
// @Security    ApiKeyAuth
// @Router      /kanbans/{boardID}/columns [put]
func (a *App) putKanbanColumns(w http.ResponseWriter, r *http.Request) {
	boardID := a.requireBoardInOrg(w, r)
	if boardID == "" {
		return
	}

	var items []putColumnItem
	if err := json.NewDecoder(r.Body).Decode(&items); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	for _, item := range items {
		if item.Name == "" || item.ZendeskStatus == "" || item.Color == "" {
			http.Error(w, "each column requires name, zendesk_status, and color", http.StatusBadRequest)
			return
		}
	}

	// Fetch existing column IDs for this board upfront so we can validate
	// that any provided IDs actually belong here before opening a transaction.
	existRows, err := a.db.QueryContext(r.Context(),
		`SELECT id FROM board_columns WHERE board_id = $1`, boardID)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("putKanbanColumns fetch existing: %v", err)
		return
	}
	existingIDs := map[string]bool{}
	for existRows.Next() {
		var id string
		if err := existRows.Scan(&id); err != nil {
			http.Error(w, "scan failed", http.StatusInternalServerError)
			log.Printf("putKanbanColumns scan existing: %v", err)
			return
		}
		existingIDs[id] = true
	}
	existRows.Close()

	// Validate all provided IDs belong to this board
	keepIDs := map[string]bool{}
	for _, item := range items {
		if item.ID == nil {
			continue
		}
		if !existingIDs[*item.ID] {
			http.Error(w, "column not found: "+*item.ID, http.StatusBadRequest)
			return
		}
		keepIDs[*item.ID] = true
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "transaction failed", http.StatusInternalServerError)
		log.Printf("putKanbanColumns begin tx: %v", err)
		return
	}
	defer tx.Rollback()

	// Delete columns absent from the request
	for id := range existingIDs {
		if keepIDs[id] {
			continue
		}
		if _, err := tx.ExecContext(r.Context(),
			`DELETE FROM board_columns WHERE id = $1`, id); err != nil {
			http.Error(w, "delete failed", http.StatusInternalServerError)
			log.Printf("putKanbanColumns delete column %s: %v", id, err)
			return
		}
	}

	// Update existing columns and insert new ones, preserving request order
	for _, item := range items {
		if item.ID != nil {
			_, err = tx.ExecContext(r.Context(), `
				UPDATE board_columns
				SET name=$2, position=$3, zendesk_status=$4::zendesk_status_category, color=$5
				WHERE id=$1
			`, *item.ID, item.Name, item.Position, item.ZendeskStatus, item.Color)
		} else {
			_, err = tx.ExecContext(r.Context(), `
				INSERT INTO board_columns (board_id, name, position, zendesk_status, color)
				VALUES ($1, $2, $3, $4::zendesk_status_category, $5)
			`, boardID, item.Name, item.Position, item.ZendeskStatus, item.Color)
		}
		if err != nil {
			http.Error(w, "write failed", http.StatusInternalServerError)
			log.Printf("putKanbanColumns write column: %v", err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "commit failed", http.StatusInternalServerError)
		log.Printf("putKanbanColumns commit: %v", err)
		return
	}

	// Return the full updated column list so the client gets IDs for new columns
	colRows, err := a.db.QueryContext(r.Context(), `
		SELECT id, created_at, updated_at, name, position, zendesk_status, color
		FROM board_columns
		WHERE board_id = $1
		ORDER BY position ASC
	`, boardID)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("putKanbanColumns fetch result: %v", err)
		return
	}
	defer colRows.Close()

	columns := []kanbanColumn{}
	for colRows.Next() {
		var c kanbanColumn
		if err := colRows.Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt, &c.Name, &c.Position, &c.ZendeskStatus, &c.Color); err != nil {
			http.Error(w, "scan failed", http.StatusInternalServerError)
			log.Printf("putKanbanColumns scan result: %v", err)
			return
		}
		columns = append(columns, c)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(columns)
}
