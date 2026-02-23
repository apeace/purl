package app

import (
	"database/sql"
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

type updateKanbanRequest struct {
	Name string `json:"name"`
}

// @Summary     Update a Kanban board
// @Description Updates the name of a Kanban board
// @Accept      json
// @Produce     json
// @Param       boardID  path      string               true  "Board ID"
// @Param       body     body      updateKanbanRequest  true  "Fields to update"
// @Success     200      {object}  kanbanBoard
// @Failure     400      {string}  string  "Bad Request"
// @Failure     401      {string}  string  "Unauthorized"
// @Failure     403      {string}  string  "Forbidden"
// @Failure     404      {string}  string  "Not Found"
// @Security    ApiKeyAuth
// @Router      /kanbans/{boardID} [patch]
func (a *App) updateKanban(w http.ResponseWriter, r *http.Request) {
	boardID := a.requireBoardInOrg(w, r)
	if boardID == "" {
		return
	}

	var req updateKanbanRequest
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
		UPDATE boards SET name = $2
		WHERE id = $1
		RETURNING id, created_at, updated_at, name, is_default
	`, boardID, req.Name).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt, &b.Name, &b.IsDefault)
	if err != nil {
		http.Error(w, "update failed", http.StatusInternalServerError)
		log.Printf("updateKanban update: %v", err)
		return
	}
	b.Columns = []kanbanColumn{}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(b)
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

// requireBoardInOrg checks that the board URL param exists, belongs to the org,
// and is not a default board (which is read-only).
// Returns the board ID on success, or writes an appropriate error and returns "".
func (a *App) requireBoardInOrg(w http.ResponseWriter, r *http.Request) string {
	boardID := chi.URLParam(r, "boardID")
	o := orgFromContext(r.Context())

	var isDefault bool
	err := a.db.QueryRowContext(r.Context(),
		`SELECT is_default FROM boards WHERE id = $1 AND org_id = $2`,
		boardID, o.ID,
	).Scan(&isDefault)
	if err == sql.ErrNoRows {
		http.Error(w, "board not found", http.StatusNotFound)
		return ""
	}
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("requireBoardInOrg query: %v", err)
		return ""
	}
	if isDefault {
		http.Error(w, "default board is read-only", http.StatusForbidden)
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

// requireColumnInBoard checks that the column URL param exists and belongs to the board.
// Returns the column ID on success, or writes a 404 and returns "".
func (a *App) requireColumnInBoard(w http.ResponseWriter, r *http.Request, boardID string) string {
	columnID := chi.URLParam(r, "columnID")

	var exists bool
	err := a.db.QueryRowContext(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM board_columns WHERE id = $1 AND board_id = $2)`,
		columnID, boardID,
	).Scan(&exists)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("requireColumnInBoard query: %v", err)
		return ""
	}
	if !exists {
		http.Error(w, "column not found", http.StatusNotFound)
		return ""
	}
	return columnID
}

// @Summary     Replace column tickets
// @Description Atomically sets the ordered list of tickets in a column.
// @Description Tickets are placed at positions matching their array index.
// @Description A ticket already in another column on this board is moved here.
// @Description Tickets previously in this column but absent from the request are removed from the board.
// @Accept      json
// @Produce     json
// @Param       boardID   path      string    true  "Board ID"
// @Param       columnID  path      string    true  "Column ID"
// @Param       body      body      []string  true  "Ordered ticket IDs"
// @Success     200       {array}   string
// @Failure     400       {string}  string  "Bad Request"
// @Failure     401       {string}  string  "Unauthorized"
// @Failure     404       {string}  string  "Not Found"
// @Security    ApiKeyAuth
// @Router      /kanbans/{boardID}/columns/{columnID}/tickets [put]
func (a *App) putColumnTickets(w http.ResponseWriter, r *http.Request) {
	boardID := a.requireBoardInOrg(w, r)
	if boardID == "" {
		return
	}
	columnID := a.requireColumnInBoard(w, r, boardID)
	if columnID == "" {
		return
	}

	var ticketIDs []string
	if err := json.NewDecoder(r.Body).Decode(&ticketIDs); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate all ticket IDs belong to this org before touching anything
	o := orgFromContext(r.Context())
	for _, tid := range ticketIDs {
		var exists bool
		err := a.db.QueryRowContext(r.Context(),
			`SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1 AND org_id = $2)`,
			tid, o.ID,
		).Scan(&exists)
		if err != nil {
			http.Error(w, "query failed", http.StatusInternalServerError)
			log.Printf("putColumnTickets validate ticket %s: %v", tid, err)
			return
		}
		if !exists {
			http.Error(w, "ticket not found: "+tid, http.StatusBadRequest)
			return
		}
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "transaction failed", http.StatusInternalServerError)
		log.Printf("putColumnTickets begin tx: %v", err)
		return
	}
	defer tx.Rollback()

	// Remove all tickets currently in this column. Tickets being moved here
	// from another column are handled per-ticket below.
	if _, err := tx.ExecContext(r.Context(),
		`DELETE FROM kanban_board_tickets WHERE column_id = $1`, columnID,
	); err != nil {
		http.Error(w, "delete failed", http.StatusInternalServerError)
		log.Printf("putColumnTickets clear column: %v", err)
		return
	}

	for i, tid := range ticketIDs {
		// If this ticket is in another column on this board, remove it from there first
		if _, err := tx.ExecContext(r.Context(),
			`DELETE FROM kanban_board_tickets WHERE board_id = $1 AND ticket_id = $2`,
			boardID, tid,
		); err != nil {
			http.Error(w, "delete failed", http.StatusInternalServerError)
			log.Printf("putColumnTickets evict ticket %s: %v", tid, err)
			return
		}

		if _, err := tx.ExecContext(r.Context(), `
			INSERT INTO kanban_board_tickets (board_id, column_id, ticket_id, position)
			VALUES ($1, $2, $3, $4)
		`, boardID, columnID, tid, i); err != nil {
			http.Error(w, "insert failed", http.StatusInternalServerError)
			log.Printf("putColumnTickets insert ticket %s: %v", tid, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "commit failed", http.StatusInternalServerError)
		log.Printf("putColumnTickets commit: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticketIDs)
}

type kanbanTicketRow struct {
	ColumnID     string    `json:"column_id"`
	Position     int       `json:"position"`
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	Priority     string    `json:"priority"`
	ReporterName string    `json:"reporter_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// @Summary     List tickets on a Kanban board
// @Description Returns all tickets placed on the board, ordered by column then position.
// @Description Each row includes column_id and position so the client can reconstruct per-column ordering.
// @Produce     json
// @Param       boardID  path      string  true  "Board ID"
// @Success     200      {array}   kanbanTicketRow
// @Failure     401      {string}  string  "Unauthorized"
// @Failure     404      {string}  string  "Not Found"
// @Security    ApiKeyAuth
// @Router      /kanbans/{boardID}/tickets [get]
func (a *App) listKanbanTickets(w http.ResponseWriter, r *http.Request) {
	boardID := chi.URLParam(r, "boardID")
	o := orgFromContext(r.Context())

	// Verify the board exists and belongs to this org (readable regardless of is_default)
	var exists bool
	err := a.db.QueryRowContext(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM boards WHERE id = $1 AND org_id = $2)`,
		boardID, o.ID,
	).Scan(&exists)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("listKanbanTickets board check: %v", err)
		return
	}
	if !exists {
		http.Error(w, "board not found", http.StatusNotFound)
		return
	}

	rows, err := a.db.QueryContext(r.Context(), `
		SELECT kbt.column_id, kbt.position, t.id, t.title, t.status, t.priority, c.name, t.created_at
		FROM kanban_board_tickets kbt
		JOIN tickets t ON t.id = kbt.ticket_id
		JOIN customers c ON c.id = t.reporter_id
		WHERE kbt.board_id = $1
		ORDER BY kbt.column_id, kbt.position ASC
	`, boardID)
	if err != nil {
		http.Error(w, "query failed", http.StatusInternalServerError)
		log.Printf("listKanbanTickets query: %v", err)
		return
	}
	defer rows.Close()

	tickets := []kanbanTicketRow{}
	for rows.Next() {
		var t kanbanTicketRow
		if err := rows.Scan(&t.ColumnID, &t.Position, &t.ID, &t.Title, &t.Status, &t.Priority, &t.ReporterName, &t.CreatedAt); err != nil {
			http.Error(w, "scan failed", http.StatusInternalServerError)
			log.Printf("listKanbanTickets scan: %v", err)
			return
		}
		tickets = append(tickets, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}
