package app

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type ZendeskTicket struct {
	ID          int64     `json:"id"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	RequesterID int64     `json:"requester_id"`
	AssigneeID  *int64    `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ZendeskTicketsResponse struct {
	Tickets  []ZendeskTicket `json:"tickets"`
	NextPage *string         `json:"next_page"`
}

type ZendeskCommentVia struct {
	Channel string `json:"channel"`
}

// zendeskVoiceData captures the structured fields Zendesk returns in the "data"
// object of VoiceComment-type comments. All fields are pointers because they're
// only present on voice comments (and some are conditional even then).
type zendeskVoiceData struct {
	CallID              *int64     `json:"call_id"`
	RecordingURL        *string    `json:"recording_url"`
	TranscriptionText   *string    `json:"transcription_text"`
	TranscriptionStatus *string    `json:"transcription_status"`
	CallDuration        *int       `json:"call_duration"`
	From                *string    `json:"from"`
	To                  *string    `json:"to"`
	FormattedFrom       *string    `json:"formatted_from"`
	FormattedTo         *string    `json:"formatted_to"`
	AnsweredByID        *int64     `json:"answered_by_id"`
	AnsweredByName      *string    `json:"answered_by_name"`
	Location            *string    `json:"location"`
	StartedAt           *time.Time `json:"started_at"`
}

func (d *zendeskVoiceData) callFrom() *string {
	if d == nil {
		return nil
	}
	if d.FormattedFrom != nil {
		return d.FormattedFrom
	}
	return d.From
}

func (d *zendeskVoiceData) callTo() *string {
	if d == nil {
		return nil
	}
	if d.FormattedTo != nil {
		return d.FormattedTo
	}
	return d.To
}

type ZendeskComment struct {
	ID        int64              `json:"id"`
	Type      string             `json:"type"`
	Body      string             `json:"body"`
	AuthorID  int64              `json:"author_id"`
	Public    bool               `json:"public"`
	Via       ZendeskCommentVia  `json:"via"`
	CreatedAt time.Time          `json:"created_at"`
	Data      *zendeskVoiceData  `json:"data"`
}

type ZendeskCommentsResponse struct {
	Comments []ZendeskComment `json:"comments"`
}

type ZendeskUser struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"` // "end-user", "agent", "admin"
}

type ZendeskUsersResponse struct {
	Users    []ZendeskUser `json:"users"`
	NextPage *string       `json:"next_page"`
}

func ZendeskGet(subdomain, creds, path string) ([]byte, error) {
	url := fmt.Sprintf("https://%s.zendesk.com%s", subdomain, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s: %s", resp.Status, body)
	}
	return body, nil
}

// fetchAllAgents retrieves all agents and admins from Zendesk, handling pagination.
func fetchAllAgents(subdomain, creds string) ([]ZendeskUser, error) {
	var all []ZendeskUser
	path := "/api/v2/users.json?role[]=agent&role[]=admin&per_page=100"

	for path != "" {
		body, err := ZendeskGet(subdomain, creds, path)
		if err != nil {
			return nil, err
		}
		var resp ZendeskUsersResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return nil, fmt.Errorf("parse users: %w", err)
		}
		all = append(all, resp.Users...)

		if resp.NextPage == nil {
			break
		}
		// NextPage is a full URL; extract the path+query portion
		next := *resp.NextPage
		idx := strings.Index(next, "/api/v2/")
		if idx == -1 {
			break
		}
		path = next[idx:]
	}

	return all, nil
}

// ImportZendeskData wipes all Zendesk-sourced data for the given org and re-imports it
// fresh from the Zendesk API. Safe to call multiple times; each call starts from a clean slate.
func ImportZendeskData(_ context.Context, db *sql.DB, orgID, subdomain, email, apiKey string) error {
	creds := base64.StdEncoding.EncodeToString([]byte(email + "/token:" + apiKey))

	// Wipe existing Zendesk data. Order matters: tickets must be deleted before customers
	// because tickets reference customers via reporter_id. Agents with a zendesk_user_id
	// are re-upserted below, so wipe them too for a clean import.
	log.Println("wiping existing Zendesk data for org...")
	if _, err := db.Exec(`DELETE FROM zendesk_webhook_events WHERE org_id = $1`, orgID); err != nil {
		return fmt.Errorf("wipe zendesk_webhook_events: %w", err)
	}
	// Cascades to ticket_comments and board_tickets.
	if _, err := db.Exec(`DELETE FROM tickets WHERE org_id = $1`, orgID); err != nil {
		return fmt.Errorf("wipe tickets: %w", err)
	}
	// Cascades to customer_emails.
	if _, err := db.Exec(`DELETE FROM customers WHERE org_id = $1`, orgID); err != nil {
		return fmt.Errorf("wipe customers: %w", err)
	}
	if _, err := db.Exec(`DELETE FROM agents WHERE org_id = $1 AND zendesk_user_id IS NOT NULL`, orgID); err != nil {
		return fmt.Errorf("wipe agents: %w", err)
	}

	// Step 1: Fetch all agents and admins from Zendesk and upsert into DB.
	log.Println("fetching all agents...")
	allAgents, err := fetchAllAgents(subdomain, creds)
	if err != nil {
		return fmt.Errorf("fetch agents: %w", err)
	}
	log.Printf("fetched %d agents", len(allAgents))

	agentsByZendeskID := make(map[int64]string)     // zendeskUserID -> purl agent UUID
	agentNamesByZendeskID := make(map[int64]string) // zendeskUserID -> display name
	for _, u := range allAgents {
		var agentID string
		err := db.QueryRow(
			`INSERT INTO agents (email, name, org_id, zendesk_user_id) VALUES ($1, $2, $3, $4)
			 ON CONFLICT (org_id, email) DO UPDATE SET name = EXCLUDED.name, zendesk_user_id = EXCLUDED.zendesk_user_id
			 RETURNING id`,
			u.Email, u.Name, orgID, u.ID,
		).Scan(&agentID)
		if err != nil {
			return fmt.Errorf("insert agent %s: %w", u.Email, err)
		}
		agentsByZendeskID[u.ID] = agentID
		agentNamesByZendeskID[u.ID] = u.Name
	}
	log.Printf("upserted %d agents", len(agentsByZendeskID))

	// Step 2: Fetch tickets
	log.Println("fetching tickets...")
	ticketsBody, err := ZendeskGet(subdomain, creds, "/api/v2/tickets.json?sort_by=created_at&sort_order=desc&per_page=50")
	if err != nil {
		return fmt.Errorf("fetch tickets: %w", err)
	}
	var ticketsResp ZendeskTicketsResponse
	if err := json.Unmarshal(ticketsBody, &ticketsResp); err != nil {
		return fmt.Errorf("parse tickets: %w", err)
	}
	log.Printf("fetched %d tickets", len(ticketsResp.Tickets))

	// Step 3: Fetch comments per ticket and collect end-user IDs
	allComments := make(map[int64][]ZendeskComment) // zendeskTicketID -> comments
	endUserIDSet := make(map[int64]bool)

	for i, ticket := range ticketsResp.Tickets {
		log.Printf("fetching comments for ticket %d/%d (Zendesk ID %d)...", i+1, len(ticketsResp.Tickets), ticket.ID)
		commentsBody, err := ZendeskGet(subdomain, creds, fmt.Sprintf("/api/v2/tickets/%d/comments.json", ticket.ID))
		if err != nil {
			return fmt.Errorf("fetch comments for ticket %d: %w", ticket.ID, err)
		}
		var commentsResp ZendeskCommentsResponse
		if err := json.Unmarshal(commentsBody, &commentsResp); err != nil {
			return fmt.Errorf("parse comments for ticket %d: %w", ticket.ID, err)
		}
		allComments[ticket.ID] = commentsResp.Comments

		endUserIDSet[ticket.RequesterID] = true
		for _, c := range commentsResp.Comments {
			if c.AuthorID <= 0 {
				continue // system user; handled separately during comment import
			}
			// Only collect IDs of end-users; agents are already fetched above
			if _, isAgent := agentsByZendeskID[c.AuthorID]; !isAgent {
				endUserIDSet[c.AuthorID] = true
			}
		}
	}

	// Step 4: Batch-fetch end-users
	endUserIDs := make([]string, 0, len(endUserIDSet))
	for id := range endUserIDSet {
		endUserIDs = append(endUserIDs, fmt.Sprintf("%d", id))
	}
	log.Printf("batch-fetching %d unique end-users...", len(endUserIDs))

	customersByZendeskID := make(map[int64]string) // zendeskUserID -> purl customer UUID

	if len(endUserIDs) > 0 {
		usersBody, err := ZendeskGet(subdomain, creds, "/api/v2/users/show_many.json?ids="+strings.Join(endUserIDs, ","))
		if err != nil {
			return fmt.Errorf("fetch users: %w", err)
		}
		var usersResp ZendeskUsersResponse
		if err := json.Unmarshal(usersBody, &usersResp); err != nil {
			return fmt.Errorf("parse users: %w", err)
		}

		// Step 5: Insert customers. Also upsert any agent-role users that were
		// missed in step 1 (e.g. bot/automation accounts, or agents beyond the
		// agents fetch limit) so their comments can be attributed correctly.
		for _, u := range usersResp.Users {
			if u.Role == "end-user" {
				var customerID string
				err := db.QueryRow(
					`INSERT INTO customers (name, org_id, zendesk_user_id) VALUES ($1, $2, $3) RETURNING id`,
					u.Name, orgID, u.ID,
				).Scan(&customerID)
				if err != nil {
					return fmt.Errorf("insert customer %s: %w", u.Email, err)
				}
				_, err = db.Exec(
					`INSERT INTO customer_emails (customer_id, email, verified) VALUES ($1, $2, false)`,
					customerID, u.Email,
				)
				if err != nil {
					return fmt.Errorf("insert customer email %s: %w", u.Email, err)
				}
				customersByZendeskID[u.ID] = customerID
			} else if _, alreadyImported := agentsByZendeskID[u.ID]; !alreadyImported {
				var agentID string
				err := db.QueryRow(
					`INSERT INTO agents (email, name, org_id, zendesk_user_id)
					 VALUES ($1, $2, $3, $4)
					 ON CONFLICT (org_id, zendesk_user_id) DO UPDATE SET name = EXCLUDED.name, email = EXCLUDED.email
					 RETURNING id`,
					u.Email, u.Name, orgID, u.ID,
				).Scan(&agentID)
				if err != nil {
					return fmt.Errorf("insert agent %s: %w", u.Name, err)
				}
				agentsByZendeskID[u.ID] = agentID
			}
		}
	}
	log.Printf("inserted %d customers", len(customersByZendeskID))

	// Step 6: Insert tickets with correct timestamps
	ticketsByZendeskID := make(map[int64]string) // zendeskTicketID -> purl ticket UUID

	for _, ticket := range ticketsResp.Tickets {
		reporterID, ok := customersByZendeskID[ticket.RequesterID]
		if !ok {
			log.Printf("warn: skipping ticket %d — requester %d not found in customers map", ticket.ID, ticket.RequesterID)
			continue
		}

		var assigneeID *string
		if ticket.AssigneeID != nil {
			if id, ok := agentsByZendeskID[*ticket.AssigneeID]; ok {
				assigneeID = &id
			}
		}

		var ticketID string
		err := db.QueryRow(
			`INSERT INTO tickets (title, description, reporter_id, assignee_id, org_id, created_at, updated_at, zendesk_status, zendesk_ticket_id)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8::zendesk_status_category, $9)
			 RETURNING id`,
			ticket.Subject,
			ticket.Description,
			reporterID,
			assigneeID,
			orgID,
			ticket.CreatedAt,
			ticket.UpdatedAt,
			mapZendeskStatus(ticket.Status),
			ticket.ID,
		).Scan(&ticketID)
		if err != nil {
			return fmt.Errorf("insert ticket %d: %w", ticket.ID, err)
		}
		ticketsByZendeskID[ticket.ID] = ticketID
	}
	log.Printf("inserted %d tickets", len(ticketsByZendeskID))

	// Step 7: Insert comments with correct timestamps
	commentsInserted := 0
	var systemAgentID string // resolved lazily on first system-authored comment
	for zendeskTicketID, comments := range allComments {
		ticketID, ok := ticketsByZendeskID[zendeskTicketID]
		if !ok {
			// Ticket was skipped (e.g. reporter not found)
			continue
		}
		for _, c := range comments {
			var customerAuthorID *string
			var agentAuthorID *string
			var role string

			if c.AuthorID <= 0 {
				// Zendesk system/automation user. Resolve lazily and cache.
				if systemAgentID == "" {
					id, err := resolveSystemAgent(context.Background(), db, orgID)
					if err != nil {
						return fmt.Errorf("resolve system agent: %w", err)
					}
					systemAgentID = id
				}
				agentAuthorID = &systemAgentID
				role = "agent"
			} else if id, ok := customersByZendeskID[c.AuthorID]; ok {
				customerAuthorID = &id
				role = "customer"
			} else if id, ok := agentsByZendeskID[c.AuthorID]; ok {
				agentAuthorID = &id
				role = "agent"
			} else {
				log.Printf("warn: skipping comment %d on ticket %d — author %d not found", c.ID, zendeskTicketID, c.AuthorID)
				continue
			}

			// Extract voice data if this is a VoiceComment
			var callID *int64
			var recordingURL, transcriptionText, transcriptionStatus *string
			var callDuration *int
			var callFrom, callTo, answeredByName, callLocation *string
			var callStartedAt *time.Time
			if c.Data != nil {
				callID = c.Data.CallID
				recordingURL = c.Data.RecordingURL
				transcriptionText = c.Data.TranscriptionText
				transcriptionStatus = c.Data.TranscriptionStatus
				callDuration = c.Data.CallDuration
				callFrom = c.Data.callFrom()
				callTo = c.Data.callTo()
				callLocation = c.Data.Location
				callStartedAt = c.Data.StartedAt
				// Resolve agent name from answered_by_id if available
				answeredByName = c.Data.AnsweredByName
				if answeredByName == nil && c.Data.AnsweredByID != nil {
					if name, ok := agentNamesByZendeskID[*c.Data.AnsweredByID]; ok {
						answeredByName = &name
					}
				}
			}

			_, err := db.Exec(
				`INSERT INTO ticket_comments
					(ticket_id, customer_author_id, agent_author_id, role, body, channel, zendesk_comment_id, created_at, updated_at,
					 call_id, recording_url, transcription_text, transcription_status, call_duration,
					 call_from, call_to, answered_by_name, call_location, call_started_at)
				 VALUES ($1, $2, $3, $4::comment_role, $5, $6::comment_channel, $7, $8, $8,
				         $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
				ticketID, customerAuthorID, agentAuthorID, role, c.Body, mapCommentChannel(c.Via.Channel, c.Public), c.ID, c.CreatedAt,
				callID, recordingURL, transcriptionText, transcriptionStatus, callDuration,
				callFrom, callTo, answeredByName, callLocation, callStartedAt,
			)
			if err != nil {
				return fmt.Errorf("insert comment %d: %w", c.ID, err)
			}
			commentsInserted++
		}
	}
	log.Printf("inserted %d comments", commentsInserted)

	// Step 8: Place tickets into the default Kanban board columns by zendesk_status.
	// Deleting tickets above cascades to board_tickets, so we start fresh.
	log.Println("placing tickets into default Kanban board...")

	var defaultBoardID string
	err = db.QueryRow(
		`SELECT id FROM boards WHERE org_id = $1 AND is_default = true`,
		orgID,
	).Scan(&defaultBoardID)
	if err == sql.ErrNoRows {
		log.Println("warn: no default Kanban board found — skipping ticket placement")
		log.Println("done")
		return nil
	}
	if err != nil {
		return fmt.Errorf("query default board: %w", err)
	}

	// Check which zendesk_status values have a matching column and warn about gaps
	colRows, err := db.Query(
		`SELECT zendesk_status::text FROM board_columns WHERE board_id = $1`,
		defaultBoardID,
	)
	if err != nil {
		return fmt.Errorf("query board columns: %w", err)
	}
	coveredStatuses := map[string]bool{}
	for colRows.Next() {
		var s string
		if err := colRows.Scan(&s); err != nil {
			return fmt.Errorf("scan board column: %w", err)
		}
		coveredStatuses[s] = true
	}
	colRows.Close()

	for _, s := range []string{"new", "open", "pending", "solved", "closed"} {
		if !coveredStatuses[s] {
			log.Printf("warn: default Kanban board has no column for zendesk_status %q — those tickets will not be placed", s)
		}
	}

	// Insert all tickets into their matching column in one query.
	// Position within each column is assigned by created_at ASC (oldest = top of queue).
	result, err := db.Exec(`
		INSERT INTO board_tickets (board_id, column_id, ticket_id, position)
		SELECT
			$1,
			bc.id,
			t.id,
			(ROW_NUMBER() OVER (PARTITION BY bc.id ORDER BY t.created_at ASC) - 1)::integer
		FROM tickets t
		JOIN board_columns bc ON bc.board_id = $1 AND bc.zendesk_status = t.zendesk_status
		WHERE t.org_id = $2
	`, defaultBoardID, orgID)
	if err != nil {
		return fmt.Errorf("place tickets in kanban: %w", err)
	}
	placed, _ := result.RowsAffected()
	log.Printf("placed %d tickets into default Kanban board", placed)
	log.Println("done")
	return nil
}
