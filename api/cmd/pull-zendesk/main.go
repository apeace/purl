package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type ZendeskTicket struct {
	ID          int64   `json:"id"`
	Subject     string  `json:"subject"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    *string `json:"priority"`
	RequesterID int64   `json:"requester_id"`
	AssigneeID  *int64  `json:"assignee_id"`
}

type ZendeskTicketsResponse struct {
	Tickets []ZendeskTicket `json:"tickets"`
}

type ZendeskComment struct {
	ID       int64  `json:"id"`
	Body     string `json:"body"`
	AuthorID int64  `json:"author_id"`
	Public   bool   `json:"public"`
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
	Users []ZendeskUser `json:"users"`
}

func zendeskGet(subdomain, creds, path string) ([]byte, error) {
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

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: pull-zendesk <org-slug>")
	}

	slug := os.Args[1]

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	var orgID, subdomain, email, apiKey string
	err = db.QueryRow(
		`SELECT id, zendesk_subdomain, COALESCE(zendesk_email, ''), COALESCE(zendesk_api_key, '') FROM organizations WHERE slug = $1`,
		slug,
	).Scan(&orgID, &subdomain, &email, &apiKey)
	if err == sql.ErrNoRows {
		log.Fatalf("no organization found with slug %q", slug)
	}
	if err != nil {
		log.Fatalf("query org: %v", err)
	}
	if subdomain == "" || email == "" || apiKey == "" {
		log.Fatalf("org %q has no Zendesk credentials configured", slug)
	}

	creds := base64.StdEncoding.EncodeToString([]byte(email + "/token:" + apiKey))

	// Step 1: Fetch tickets
	log.Println("fetching tickets...")
	ticketsBody, err := zendeskGet(subdomain, creds, "/api/v2/tickets.json?sort_by=created_at&sort_order=desc&per_page=100")
	if err != nil {
		log.Fatalf("fetch tickets: %v", err)
	}
	var ticketsResp ZendeskTicketsResponse
	if err := json.Unmarshal(ticketsBody, &ticketsResp); err != nil {
		log.Fatalf("parse tickets: %v", err)
	}
	log.Printf("fetched %d tickets", len(ticketsResp.Tickets))

	// Step 2: Fetch comments per ticket and collect user IDs
	allComments := make(map[int64][]ZendeskComment) // zendeskTicketID -> comments
	userIDSet := make(map[int64]bool)

	for i, ticket := range ticketsResp.Tickets {
		log.Printf("fetching comments for ticket %d/%d (Zendesk ID %d)...", i+1, len(ticketsResp.Tickets), ticket.ID)
		commentsBody, err := zendeskGet(subdomain, creds, fmt.Sprintf("/api/v2/tickets/%d/comments.json", ticket.ID))
		if err != nil {
			log.Fatalf("fetch comments for ticket %d: %v", ticket.ID, err)
		}
		var commentsResp ZendeskCommentsResponse
		if err := json.Unmarshal(commentsBody, &commentsResp); err != nil {
			log.Fatalf("parse comments for ticket %d: %v", ticket.ID, err)
		}
		allComments[ticket.ID] = commentsResp.Comments

		userIDSet[ticket.RequesterID] = true
		if ticket.AssigneeID != nil {
			userIDSet[*ticket.AssigneeID] = true
		}
		for _, c := range commentsResp.Comments {
			userIDSet[c.AuthorID] = true
		}
	}

	// Step 3: Batch-fetch users
	userIDs := make([]string, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, fmt.Sprintf("%d", id))
	}
	log.Printf("batch-fetching %d unique users...", len(userIDs))
	usersBody, err := zendeskGet(subdomain, creds, "/api/v2/users/show_many.json?ids="+strings.Join(userIDs, ","))
	if err != nil {
		log.Fatalf("fetch users: %v", err)
	}
	var usersResp ZendeskUsersResponse
	if err := json.Unmarshal(usersBody, &usersResp); err != nil {
		log.Fatalf("parse users: %v", err)
	}

	// Step 4: Insert users into DB
	customersByZendeskID := make(map[int64]string) // zendeskUserID -> purl customer UUID
	agentsByZendeskID := make(map[int64]string)    // zendeskUserID -> purl agent UUID

	customersInserted := 0
	agentsInserted := 0

	for _, u := range usersResp.Users {
		switch u.Role {
		case "end-user":
			var customerID string
			err := db.QueryRow(
				`INSERT INTO customers (name, org_id) VALUES ($1, $2) RETURNING id`,
				u.Name, orgID,
			).Scan(&customerID)
			if err != nil {
				log.Fatalf("insert customer %s: %v", u.Email, err)
			}
			_, err = db.Exec(
				`INSERT INTO customer_emails (customer_id, email, verified) VALUES ($1, $2, false)`,
				customerID, u.Email,
			)
			if err != nil {
				log.Fatalf("insert customer email %s: %v", u.Email, err)
			}
			customersByZendeskID[u.ID] = customerID
			customersInserted++

		case "agent", "admin":
			var agentID string
			err := db.QueryRow(
				`INSERT INTO agents (email, name, org_id) VALUES ($1, $2, $3)
				 ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name
				 RETURNING id`,
				u.Email, u.Name, orgID,
			).Scan(&agentID)
			if err != nil {
				log.Fatalf("insert agent %s: %v", u.Email, err)
			}
			agentsByZendeskID[u.ID] = agentID
			agentsInserted++
		}
	}
	log.Printf("inserted %d customers, %d agents", customersInserted, agentsInserted)

	// Step 5: Insert tickets
	ticketsByZendeskID := make(map[int64]string) // zendeskTicketID -> purl ticket UUID

	for _, ticket := range ticketsResp.Tickets {
		status := mapStatus(ticket.Status)
		priority := mapPriority(ticket.Priority)

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
			`INSERT INTO tickets (title, description, status, priority, reporter_id, assignee_id, org_id)
			 VALUES ($1, $2, $3::ticket_status, $4::ticket_priority, $5, $6, $7)
			 RETURNING id`,
			ticket.Subject,
			ticket.Description,
			status,
			priority,
			reporterID,
			assigneeID,
			orgID,
		).Scan(&ticketID)
		if err != nil {
			log.Fatalf("insert ticket %d: %v", ticket.ID, err)
		}
		ticketsByZendeskID[ticket.ID] = ticketID
	}
	log.Printf("inserted %d tickets", len(ticketsByZendeskID))

	// Step 6: Insert comments
	commentsInserted := 0
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

			if id, ok := customersByZendeskID[c.AuthorID]; ok {
				customerAuthorID = &id
				role = "customer"
			} else if id, ok := agentsByZendeskID[c.AuthorID]; ok {
				agentAuthorID = &id
				role = "agent"
			} else {
				log.Printf("warn: skipping comment %d on ticket %d — author %d not found", c.ID, zendeskTicketID, c.AuthorID)
				continue
			}

			_, err := db.Exec(
				`INSERT INTO comments (ticket_id, customer_author_id, agent_author_id, role, body)
				 VALUES ($1, $2, $3, $4::comment_role, $5)`,
				ticketID, customerAuthorID, agentAuthorID, role, c.Body,
			)
			if err != nil {
				log.Fatalf("insert comment %d: %v", c.ID, err)
			}
			commentsInserted++
		}
	}
	log.Printf("inserted %d comments", commentsInserted)
	log.Println("done")
}

func mapStatus(s string) string {
	switch s {
	case "new", "open":
		return "open"
	case "pending", "hold":
		return "in_progress"
	case "solved":
		return "resolved"
	case "closed":
		return "closed"
	default:
		return "open"
	}
}

func mapPriority(p *string) string {
	if p == nil {
		return "medium"
	}
	switch *p {
	case "low":
		return "low"
	case "high":
		return "high"
	case "urgent":
		return "urgent"
	default:
		return "medium"
	}
}
