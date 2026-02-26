package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"purl/api/internal/app"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: zd-ticket <org-slug> <zendesk-ticket-number> [--json]")
	}

	slug := os.Args[1]
	ticketNumber, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		log.Fatalf("invalid ticket number %q: %v", os.Args[2], err)
	}

	rawJSON := len(os.Args) > 3 && os.Args[3] == "--json"

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

	var subdomain, email, apiKey string
	err = db.QueryRow(
		`SELECT zendesk_subdomain, COALESCE(zendesk_email, ''), COALESCE(zendesk_api_key, '') FROM organizations WHERE slug = $1`,
		slug,
	).Scan(&subdomain, &email, &apiKey)
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

	// Fetch full ticket
	ticketBody, err := app.ZendeskGet(subdomain, creds, fmt.Sprintf("/api/v2/tickets/%d.json", ticketNumber))
	if err != nil {
		log.Fatalf("fetch ticket: %v", err)
	}

	var ticketResp struct {
		Ticket json.RawMessage `json:"ticket"`
	}
	if err := json.Unmarshal(ticketBody, &ticketResp); err != nil {
		log.Fatalf("parse ticket response: %v", err)
	}

	if rawJSON {
		// Also fetch comments and output everything as JSON
		commentsBody, err := app.ZendeskGet(subdomain, creds, fmt.Sprintf("/api/v2/tickets/%d/comments.json", ticketNumber))
		if err != nil {
			log.Fatalf("fetch comments: %v", err)
		}
		var commentsResp json.RawMessage
		if err := json.Unmarshal(commentsBody, &commentsResp); err != nil {
			log.Fatalf("parse comments response: %v", err)
		}

		combined := map[string]json.RawMessage{
			"ticket":   ticketResp.Ticket,
			"comments": commentsResp,
		}
		out, _ := json.MarshalIndent(combined, "", "  ")
		fmt.Println(string(out))
		return
	}

	// Parse ticket for formatted output
	var ticket struct {
		ID               int64            `json:"id"`
		Subject          string           `json:"subject"`
		Description      string           `json:"description"`
		Status           string           `json:"status"`
		Priority         *string          `json:"priority"`
		Type             *string          `json:"type"`
		RequesterID      int64            `json:"requester_id"`
		AssigneeID       *int64           `json:"assignee_id"`
		Tags             []string         `json:"tags"`
		CustomFields     []customField    `json:"custom_fields"`
		Via              viaDetail        `json:"via"`
		SatisfactionRating json.RawMessage `json:"satisfaction_rating"`
		CreatedAt        time.Time        `json:"created_at"`
		UpdatedAt        time.Time        `json:"updated_at"`
	}
	if err := json.Unmarshal(ticketResp.Ticket, &ticket); err != nil {
		log.Fatalf("parse ticket: %v", err)
	}

	// Fetch requester
	requesterBody, err := app.ZendeskGet(subdomain, creds, fmt.Sprintf("/api/v2/users/%d.json", ticket.RequesterID))
	if err != nil {
		log.Fatalf("fetch requester: %v", err)
	}
	requester := parseUser(requesterBody)

	// Fetch assignee
	var assignee *zdUser
	if ticket.AssigneeID != nil {
		assigneeBody, err := app.ZendeskGet(subdomain, creds, fmt.Sprintf("/api/v2/users/%d.json", *ticket.AssigneeID))
		if err != nil {
			log.Fatalf("fetch assignee: %v", err)
		}
		a := parseUser(assigneeBody)
		assignee = &a
	}

	// Fetch comments
	commentsBody, err := app.ZendeskGet(subdomain, creds, fmt.Sprintf("/api/v2/tickets/%d/comments.json", ticketNumber))
	if err != nil {
		log.Fatalf("fetch comments: %v", err)
	}
	var commentsResp struct {
		Comments []struct {
			ID        int64     `json:"id"`
			AuthorID  int64     `json:"author_id"`
			Body      string    `json:"body"`
			Public    bool      `json:"public"`
			Via       viaDetail `json:"via"`
			CreatedAt time.Time `json:"created_at"`
		} `json:"comments"`
	}
	if err := json.Unmarshal(commentsBody, &commentsResp); err != nil {
		log.Fatalf("parse comments: %v", err)
	}

	// Print formatted output
	fmt.Printf("═══ Zendesk Ticket #%d ═══\n\n", ticket.ID)
	fmt.Printf("Subject:    %s\n", ticket.Subject)
	fmt.Printf("Status:     %s\n", ticket.Status)
	fmt.Printf("Priority:   %s\n", ptrOr(ticket.Priority, "—"))
	fmt.Printf("Type:       %s\n", ptrOr(ticket.Type, "—"))
	fmt.Printf("Channel:    %s\n", ticket.Via.Channel)
	fmt.Printf("Created:    %s\n", ticket.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated:    %s\n", ticket.UpdatedAt.Format(time.RFC3339))

	fmt.Printf("\n── People ──\n")
	fmt.Printf("Requester:  %s <%s>\n", requester.Name, requester.Email)
	if assignee != nil {
		fmt.Printf("Assignee:   %s <%s>\n", assignee.Name, assignee.Email)
	} else {
		fmt.Printf("Assignee:   (unassigned)\n")
	}

	if len(ticket.Tags) > 0 {
		fmt.Printf("\n── Tags ──\n")
		fmt.Printf("%s\n", strings.Join(ticket.Tags, ", "))
	}

	if len(ticket.CustomFields) > 0 {
		fmt.Printf("\n── Custom Fields ──\n")
		for _, cf := range ticket.CustomFields {
			if cf.Value == nil {
				continue
			}
			fmt.Printf("  %d: %v\n", cf.ID, cf.Value)
		}
	}

	if ticket.SatisfactionRating != nil && string(ticket.SatisfactionRating) != "null" {
		fmt.Printf("\n── CSAT ──\n")
		fmt.Printf("%s\n", string(ticket.SatisfactionRating))
	}

	fmt.Printf("\n── Description ──\n")
	desc := ticket.Description
	if len(desc) > 500 {
		desc = desc[:500] + "…"
	}
	fmt.Println(desc)

	fmt.Printf("\n── Comments (%d) ──\n", len(commentsResp.Comments))
	for i, c := range commentsResp.Comments {
		vis := "public"
		if !c.Public {
			vis = "internal"
		}
		body := c.Body
		if len(body) > 200 {
			body = body[:200] + "…"
		}
		fmt.Printf("\n  [%d] #%d by user %d (%s, via %s) at %s\n",
			i+1, c.ID, c.AuthorID, vis, c.Via.Channel, c.CreatedAt.Format(time.RFC3339))
		// Indent body lines
		for _, line := range strings.Split(body, "\n") {
			fmt.Printf("      %s\n", line)
		}
	}
}

type customField struct {
	ID    int64       `json:"id"`
	Value interface{} `json:"value"`
}

type viaDetail struct {
	Channel string `json:"channel"`
}

type zdUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func parseUser(body []byte) zdUser {
	var resp struct {
		User zdUser `json:"user"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		log.Fatalf("parse user: %v", err)
	}
	return resp.User
}

func ptrOr(p *string, fallback string) string {
	if p != nil {
		return *p
	}
	return fallback
}
