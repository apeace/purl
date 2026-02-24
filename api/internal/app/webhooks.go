package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// ── Zendesk webhook payload types ────────────────────────────────────────────

type zendeskWebhookEnvelope struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	AccountID int64           `json:"account_id"`
	Detail    json.RawMessage `json:"detail"`
}

type webhookTicketDetail struct {
	ID          int64     `json:"id"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	RequesterID int64     `json:"requester_id"`
	AssigneeID  *int64    `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type webhookTicketDeletedDetail struct {
	ID int64 `json:"id"`
}

type webhookCommentVia struct {
	Channel string `json:"channel"`
}

type webhookCommentDetail struct {
	ID        int64             `json:"id"`
	TicketID  int64             `json:"ticket_id"`
	AuthorID  int64             `json:"author_id"`
	Body      string            `json:"body"`
	Public    bool              `json:"public"`
	Via       webhookCommentVia `json:"via"`
	CreatedAt time.Time         `json:"created_at"`
}

type webhookUserDetail struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"` // "end-user", "agent", "admin"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ── Main handler ──────────────────────────────────────────────────────────────

// @Summary     Zendesk webhook receiver
// @Tags        Webhooks
// @Description Receives and processes Zendesk event subscription webhooks
// @Accept      json
// @Param       orgSlug  path  string  true  "Organization slug"
// @Success     204
// @Failure     400  {string}  string  "Bad Request"
// @Failure     401  {string}  string  "Unauthorized"
// @Failure     404  {string}  string  "Not Found"
// @Router      /webhooks/zendesk/{orgSlug} [post]
func (a *App) handleZendeskWebhook(w http.ResponseWriter, r *http.Request) {
	orgSlug := chi.URLParam(r, "orgSlug")

	// Read and buffer the body before any DB work; we need it for signature verification.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}

	// Load org and webhook secret.
	var orgID, webhookSecret string
	err = a.db.QueryRowContext(r.Context(),
		`SELECT id, COALESCE(zendesk_webhook_secret, '') FROM organizations WHERE slug = $1`,
		orgSlug,
	).Scan(&orgID, &webhookSecret)
	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		log.Printf("webhook: query org %q: %v", orgSlug, err)
		return
	}
	if webhookSecret == "" {
		http.Error(w, "webhook not configured", http.StatusUnauthorized)
		return
	}

	// Verify HMAC-SHA256 signature.
	// Zendesk signs: HMAC-SHA256(signing_secret, timestamp + body), base64-encoded.
	timestamp := r.Header.Get("X-Zendesk-Webhook-Signature-Timestamp")
	signature := r.Header.Get("X-Zendesk-Webhook-Signature")
	if !verifyZendeskSignature(webhookSecret, timestamp, body, signature) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	var envelope zendeskWebhookEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	var handlerErr error

	switch envelope.Type {
	case "zen:event-type:ticket.created", "zen:event-type:ticket.updated":
		var detail webhookTicketDetail
		if err := json.Unmarshal(envelope.Detail, &detail); err != nil {
			http.Error(w, "invalid detail", http.StatusBadRequest)
			return
		}
		handlerErr = a.handleTicketUpsert(ctx, orgID, &detail)

	case "zen:event-type:ticket.deleted":
		var detail webhookTicketDeletedDetail
		if err := json.Unmarshal(envelope.Detail, &detail); err != nil {
			http.Error(w, "invalid detail", http.StatusBadRequest)
			return
		}
		handlerErr = a.handleTicketDeleted(ctx, orgID, detail.ID)

	case "zen:event-type:comment.created":
		var detail webhookCommentDetail
		if err := json.Unmarshal(envelope.Detail, &detail); err != nil {
			http.Error(w, "invalid detail", http.StatusBadRequest)
			return
		}
		handlerErr = a.handleCommentCreated(ctx, orgID, &detail)

	case "zen:event-type:comment.updated":
		var detail webhookCommentDetail
		if err := json.Unmarshal(envelope.Detail, &detail); err != nil {
			http.Error(w, "invalid detail", http.StatusBadRequest)
			return
		}
		handlerErr = a.handleCommentUpdated(ctx, orgID, &detail)

	case "zen:event-type:user.created", "zen:event-type:user.updated":
		var detail webhookUserDetail
		if err := json.Unmarshal(envelope.Detail, &detail); err != nil {
			http.Error(w, "invalid detail", http.StatusBadRequest)
			return
		}
		handlerErr = a.handleUserUpsert(ctx, orgID, &detail)

	default:
		// Unknown event type — acknowledge and ignore.
		log.Printf("webhook: unknown event type %q for org %q", envelope.Type, orgSlug)
	}

	if handlerErr != nil {
		http.Error(w, "handler error", http.StatusInternalServerError)
		log.Printf("webhook: %s for org %q: %v", envelope.Type, orgSlug, handlerErr)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── Event handlers ────────────────────────────────────────────────────────────

func (a *App) handleTicketUpsert(ctx context.Context, orgID string, d *webhookTicketDetail) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Resolve reporter customer. If missing, fetch from Zendesk and upsert.
	reporterID, err := a.resolveOrFetchCustomer(ctx, tx, orgID, d.RequesterID)
	if err != nil {
		return fmt.Errorf("resolve reporter: %w", err)
	}
	if reporterID == "" {
		log.Printf("webhook ticket %d: reporter %d not found and could not be fetched — skipping", d.ID, d.RequesterID)
		return nil
	}

	// Resolve optional assignee agent.
	var assigneeID *string
	if d.AssigneeID != nil {
		var agentID string
		err := tx.QueryRowContext(ctx,
			`SELECT id FROM agents WHERE org_id = $1 AND zendesk_user_id = $2`,
			orgID, *d.AssigneeID,
		).Scan(&agentID)
		if err == nil {
			assigneeID = &agentID
		}
		// Unknown assignee is non-fatal; leave assignee_id NULL.
	}

	// Capture old zendesk_status before upserting, so we can detect changes.
	var oldStatus *string
	_ = tx.QueryRowContext(ctx,
		`SELECT zendesk_status::text FROM tickets WHERE org_id = $1 AND zendesk_ticket_id = $2`,
		orgID, d.ID,
	).Scan(&oldStatus)

	newStatus := mapZendeskStatus(d.Status)

	var ticketID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO tickets (title, description, reporter_id, assignee_id, org_id,
		                     zendesk_status, zendesk_ticket_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::zendesk_status_category, $7, $8, $9)
		ON CONFLICT (org_id, zendesk_ticket_id) DO UPDATE SET
			title       = EXCLUDED.title,
			description = EXCLUDED.description,
			reporter_id = EXCLUDED.reporter_id,
			assignee_id = EXCLUDED.assignee_id,
			zendesk_status = EXCLUDED.zendesk_status,
			updated_at  = EXCLUDED.updated_at
		RETURNING id`,
		d.Subject, d.Description, reporterID, assigneeID, orgID,
		newStatus, d.ID, d.CreatedAt, d.UpdatedAt,
	).Scan(&ticketID)
	if err != nil {
		return fmt.Errorf("upsert ticket: %w", err)
	}

	// Sync kanban only when the ticket is new or its status changed.
	isNew := oldStatus == nil
	statusChanged := oldStatus != nil && *oldStatus != newStatus
	if isNew || statusChanged {
		if err := syncTicketToDefaultKanban(ctx, tx, orgID, ticketID, newStatus); err != nil {
			return fmt.Errorf("sync kanban: %w", err)
		}
	}

	return tx.Commit()
}

func (a *App) handleTicketDeleted(ctx context.Context, orgID string, zendeskTicketID int64) error {
	_, err := a.db.ExecContext(ctx,
		`DELETE FROM tickets WHERE org_id = $1 AND zendesk_ticket_id = $2`,
		orgID, zendeskTicketID,
	)
	return err
}

func (a *App) handleCommentCreated(ctx context.Context, orgID string, d *webhookCommentDetail) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Resolve the ticket by its Zendesk ID.
	var ticketID string
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM tickets WHERE org_id = $1 AND zendesk_ticket_id = $2`,
		orgID, d.TicketID,
	).Scan(&ticketID)
	if err == sql.ErrNoRows {
		log.Printf("webhook comment %d: ticket %d not found — skipping", d.ID, d.TicketID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("look up ticket: %w", err)
	}

	// Resolve the author — may be a customer or an agent.
	var customerAuthorID *string
	var agentAuthorID *string
	var role string

	var customerID string
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM customers WHERE org_id = $1 AND zendesk_user_id = $2`,
		orgID, d.AuthorID,
	).Scan(&customerID)
	if err == nil {
		customerAuthorID = &customerID
		role = "customer"
	} else {
		var agentID string
		err = tx.QueryRowContext(ctx,
			`SELECT id FROM agents WHERE org_id = $1 AND zendesk_user_id = $2`,
			orgID, d.AuthorID,
		).Scan(&agentID)
		if err == nil {
			agentAuthorID = &agentID
			role = "agent"
		} else {
			// Author unknown — try fetching from Zendesk before giving up.
			fetched, fetchErr := a.fetchZendeskUser(ctx, orgID, d.AuthorID)
			if fetchErr != nil {
				log.Printf("webhook comment %d: author %d not found and fetch failed: %v — skipping", d.ID, d.AuthorID, fetchErr)
				return nil
			}
			if fetched == nil {
				log.Printf("webhook comment %d: author %d not found — skipping", d.ID, d.AuthorID)
				return nil
			}
			if fetched.Role == "end-user" {
				cid, upsertErr := upsertCustomer(ctx, tx, orgID, fetched.ID, fetched.Name, fetched.Email)
				if upsertErr != nil {
					return fmt.Errorf("upsert author customer: %w", upsertErr)
				}
				customerAuthorID = &cid
				role = "customer"
			} else {
				aid, upsertErr := upsertAgent(ctx, tx, orgID, fetched.ID, fetched.Name, fetched.Email)
				if upsertErr != nil {
					return fmt.Errorf("upsert author agent: %w", upsertErr)
				}
				agentAuthorID = &aid
				role = "agent"
			}
		}
	}

	channel := mapCommentChannel(d.Via.Channel, d.Public)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO ticket_comments
			(ticket_id, customer_author_id, agent_author_id, role, body, channel, zendesk_comment_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4::comment_role, $5, $6::comment_channel, $7, $8, $8)
		ON CONFLICT (ticket_id, zendesk_comment_id) DO NOTHING`,
		ticketID, customerAuthorID, agentAuthorID, role, d.Body, channel, d.ID, d.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert comment: %w", err)
	}

	return tx.Commit()
}

func (a *App) handleCommentUpdated(ctx context.Context, orgID string, d *webhookCommentDetail) error {
	// Only the body changes on a comment update (agent redaction).
	_, err := a.db.ExecContext(ctx, `
		UPDATE ticket_comments
		SET body = $1, updated_at = now()
		WHERE zendesk_comment_id = $2
		  AND ticket_id IN (SELECT id FROM tickets WHERE org_id = $3 AND zendesk_ticket_id = $4)`,
		d.Body, d.ID, orgID, d.TicketID,
	)
	return err
}

func (a *App) handleUserUpsert(ctx context.Context, orgID string, d *webhookUserDetail) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if d.Role == "end-user" {
		if _, err := upsertCustomer(ctx, tx, orgID, d.ID, d.Name, d.Email); err != nil {
			return fmt.Errorf("upsert customer: %w", err)
		}
	} else {
		// agent or admin
		if _, err := upsertAgent(ctx, tx, orgID, d.ID, d.Name, d.Email); err != nil {
			return fmt.Errorf("upsert agent: %w", err)
		}
	}

	return tx.Commit()
}

// ── Shared helpers ────────────────────────────────────────────────────────────

// resolveOrFetchCustomer returns the purl customer ID for a given Zendesk user ID.
// If not found in the database, it fetches the user from the Zendesk API and upserts them.
func (a *App) resolveOrFetchCustomer(ctx context.Context, tx *sql.Tx, orgID string, zendeskUserID int64) (string, error) {
	var customerID string
	err := tx.QueryRowContext(ctx,
		`SELECT id FROM customers WHERE org_id = $1 AND zendesk_user_id = $2`,
		orgID, zendeskUserID,
	).Scan(&customerID)
	if err == nil {
		return customerID, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("look up customer: %w", err)
	}

	// Not in DB — fetch from Zendesk.
	user, err := a.fetchZendeskUser(ctx, orgID, zendeskUserID)
	if err != nil {
		return "", fmt.Errorf("fetch zendesk user %d: %w", zendeskUserID, err)
	}
	if user == nil || user.Role != "end-user" {
		return "", nil // not a customer; caller will skip the ticket
	}

	id, err := upsertCustomer(ctx, tx, orgID, user.ID, user.Name, user.Email)
	if err != nil {
		return "", fmt.Errorf("upsert fetched customer: %w", err)
	}
	return id, nil
}

// upsertCustomer inserts or updates a customer row identified by (org_id, zendesk_user_id).
// Also adds the email to customer_emails if not already present.
func upsertCustomer(ctx context.Context, tx *sql.Tx, orgID string, zendeskUserID int64, name, email string) (string, error) {
	var customerID string
	err := tx.QueryRowContext(ctx, `
		INSERT INTO customers (name, org_id, zendesk_user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (org_id, zendesk_user_id) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`,
		name, orgID, zendeskUserID,
	).Scan(&customerID)
	if err != nil {
		return "", err
	}

	if email != "" {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO customer_emails (customer_id, email, verified)
			SELECT $1, $2, false
			WHERE NOT EXISTS (
				SELECT 1 FROM customer_emails WHERE customer_id = $1 AND email = $2
			)`,
			customerID, email,
		)
		if err != nil {
			return "", fmt.Errorf("upsert email: %w", err)
		}
	}

	return customerID, nil
}

// upsertAgent inserts or updates an agent row identified by (org_id, zendesk_user_id).
func upsertAgent(ctx context.Context, tx *sql.Tx, orgID string, zendeskUserID int64, name, email string) (string, error) {
	var agentID string
	err := tx.QueryRowContext(ctx, `
		INSERT INTO agents (email, name, org_id, zendesk_user_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (org_id, zendesk_user_id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email
		RETURNING id`,
		email, name, orgID, zendeskUserID,
	).Scan(&agentID)
	return agentID, err
}

// syncTicketToDefaultKanban removes a ticket from its current column in the default
// board and re-inserts it into the column matching the given zendesk_status.
// If no matching column exists, the ticket is simply removed from the board.
func syncTicketToDefaultKanban(ctx context.Context, tx *sql.Tx, orgID, ticketID, status string) error {
	var defaultBoardID string
	err := tx.QueryRowContext(ctx,
		`SELECT id FROM boards WHERE org_id = $1 AND is_default = true`,
		orgID,
	).Scan(&defaultBoardID)
	if err == sql.ErrNoRows {
		return nil // no default board configured
	}
	if err != nil {
		return fmt.Errorf("query default board: %w", err)
	}

	// Remove from any column in the default board.
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM board_tickets WHERE board_id = $1 AND ticket_id = $2`,
		defaultBoardID, ticketID,
	); err != nil {
		return fmt.Errorf("remove from board: %w", err)
	}

	// Find the column for the new status.
	var columnID string
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM board_columns WHERE board_id = $1 AND zendesk_status = $2::zendesk_status_category`,
		defaultBoardID, status,
	).Scan(&columnID)
	if err == sql.ErrNoRows {
		return nil // status not covered by any column; ticket stays off the board
	}
	if err != nil {
		return fmt.Errorf("query column for status %q: %w", status, err)
	}

	// Append at the end of the column.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO board_tickets (board_id, column_id, ticket_id, position)
		VALUES ($1, $2, $3,
			(SELECT COALESCE(MAX(position) + 1, 0) FROM board_tickets WHERE column_id = $2))`,
		defaultBoardID, columnID, ticketID,
	)
	if err != nil {
		return fmt.Errorf("insert into board: %w", err)
	}

	return nil
}

// fetchZendeskUser fetches a single user from the Zendesk REST API using the
// org's stored credentials. Returns nil if credentials are not configured.
func (a *App) fetchZendeskUser(ctx context.Context, orgID string, zendeskUserID int64) (*webhookUserDetail, error) {
	var subdomain, email, apiKey string
	err := a.db.QueryRowContext(ctx,
		`SELECT COALESCE(zendesk_subdomain,''), COALESCE(zendesk_email,''), COALESCE(zendesk_api_key,'')
		 FROM organizations WHERE id = $1`,
		orgID,
	).Scan(&subdomain, &email, &apiKey)
	if err != nil {
		return nil, fmt.Errorf("load zendesk creds: %w", err)
	}
	if subdomain == "" || email == "" || apiKey == "" {
		return nil, nil // no credentials; caller handles gracefully
	}

	creds := base64.StdEncoding.EncodeToString([]byte(email + "/token:" + apiKey))
	url := fmt.Sprintf("https://%s.zendesk.com/api/v2/users/%d.json", subdomain, zendeskUserID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+creds)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zendesk request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("zendesk returned %s: %s", resp.Status, body)
	}

	var wrapper struct {
		User webhookUserDetail `json:"user"`
	}
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("parse user: %w", err)
	}
	return &wrapper.User, nil
}

// verifyZendeskSignature validates the HMAC-SHA256 signature Zendesk attaches to
// every webhook request. The signed message is the timestamp concatenated with the
// raw request body; the key is the per-org signing secret.
func verifyZendeskSignature(secret, timestamp string, body []byte, signature string) bool {
	if secret == "" || timestamp == "" || signature == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// mapZendeskStatus maps a raw Zendesk ticket status string to our
// zendesk_status_category enum value.
func mapZendeskStatus(s string) string {
	switch s {
	case "new", "open", "pending", "solved", "closed":
		return s
	case "hold":
		return "pending"
	default:
		return "open"
	}
}

// mapCommentChannel maps a Zendesk via.channel value and public flag to our
// comment_channel enum. Private comments are always "internal".
func mapCommentChannel(viaChannel string, public bool) string {
	if !public {
		return "internal"
	}
	switch viaChannel {
	case "email":
		return "email"
	case "sms", "native_messaging", "whatsapp":
		return "sms"
	case "voice", "phone":
		return "voice"
	default:
		return "web"
	}
}
