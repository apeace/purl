package main

import (
	"database/sql"
	"log"
	"math/rand"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var customerData = []struct {
	name   string
	emails []string
	phones []string
}{
	{"Alice Chen", []string{"alice@example.com"}, []string{"555-0101"}},
	{"Bob Harris", []string{"bob@example.com", "bobby.h@gmail.com"}, []string{"555-0102"}},
	{"Carol Singh", []string{"carol@example.com"}, []string{"555-0103", "555-0203"}},
	{"Dan Okafor", []string{"dan@example.com", "dan.o@work.com"}, []string{"555-0104"}},
	{"Eve Nakamura", []string{"eve@example.com"}, []string{"555-0105"}},
	{"Frank Torres", []string{"frank@example.com"}, []string{"555-0106", "555-0206"}},
	{"Grace Müller", []string{"grace@example.com"}, []string{"555-0107"}},
	{"Henry Park", []string{"henry@example.com"}, []string{"555-0108"}},
	{"Isla Rivera", []string{"isla@example.com", "isla.r@gmail.com"}, []string{"555-0109"}},
	{"James Kowalski", []string{"james@example.com"}, []string{"555-0110"}},
}

var agentData = []struct{ email, name string }{
	{"sarah.support@brightwave.com", "Sarah Mitchell"},
	{"tom.tech@brightwave.com", "Tom Bradley"},
	{"lisa.billing@brightwave.com", "Lisa Nguyen"},
	{"mike.network@brightwave.com", "Mike Osei"},
}

var titles = []string{
	"Internet speed much slower than advertised",
	"No internet connection since yesterday evening",
	"Wi-Fi dropping every few hours",
	"Router keeps rebooting on its own",
	"Billed twice for the same month",
	"Charged for equipment I already returned",
	"Promotional rate not applied to my account",
	"Can't complete plan upgrade online — page errors",
	"Downgrade request ignored after 2 weeks",
	"Service outage with no status update from provider",
	"Technician missed scheduled appointment",
	"Upload speeds fine but download is throttled",
	"Latency spikes during peak evening hours",
	"Static IP address keeps changing",
	"Modem not recognized after firmware update",
	"New customer discount not reflected on bill",
	"Contract cancellation fee charged incorrectly",
	"Auto-pay failed but no notification sent",
	"Bundle discount removed without notice",
	"Data cap overage charge on unlimited plan",
	"IPv6 not working after plan switch",
	"DNS resolution intermittently failing",
	"Packet loss on wired connection",
	"Service outage affecting entire neighborhood",
	"Speed test shows 10 Mbps on 500 Mbps plan",
	"Unable to reach support — hold times over 2 hours",
	"Upgrade to gigabit plan not activated after payment",
	"Parental controls reset after router reboot",
	"Wrong equipment sent for self-install",
	"Final bill sent after cancellation is incorrect",
}

var descriptions = []string{
	"Started experiencing this issue two days ago. No changes on my end.",
	"Called in previously but the problem came back after a few hours.",
	"Affecting multiple devices — ruled out the router as the cause.",
	"Speed test results attached. Consistently well below the plan rate.",
	"This has been an ongoing problem for over a week with no resolution.",
	"Issue appears to be isolated to evenings between 7–10 PM.",
	"Already rebooted the modem and router multiple times. No improvement.",
	"Bill shows a charge that was not on the previous month's statement.",
}

// Weighted toward active statuses to reflect a realistic backlog.
var zendeskStatuses = []string{
	"new", "new",
	"open", "open", "open",
	"pending",
	"solved",
	"closed",
}

// Weighted toward medium.
var priorities = []string{
	"low",
	"medium", "medium", "medium",
	"high",
	"urgent",
}

var customerCommentBodies = []string{
	"Still experiencing this issue. Please advise.",
	"I rebooted the modem again — no change.",
	"This is really affecting my work-from-home setup.",
	"When can I expect this to be resolved?",
	"I've been a customer for 5 years and never had this problem before.",
	"The issue came back after your tech left.",
	"I ran another speed test just now — still well below plan speed.",
	"Can someone call me back instead? I'm having trouble with the chat.",
}

var agentCommentBodies = []string{
	"Confirmed the issue on our end. Escalating to the network team.",
	"Checked the line diagnostics — seeing elevated signal noise.",
	"Scheduled a technician visit for the next available slot.",
	"Customer confirmed the technician resolved the issue on-site.",
	"Applied a credit for the days of degraded service.",
	"This appears to be related to the area outage reported yesterday.",
	"Billing adjustment processed — should reflect on next statement.",
	"Modem logs reviewed — firmware update is recommended.",
	"Issue reproduced on our monitoring tools. Ticket forwarded to NOC.",
	"Customer confirmed service restored after the node replacement.",
	"Still investigating — provisioning team has been looped in.",
	"Plan change is now active. Speed increase should be immediate.",
}

func pick(s []string) string { return s[rand.Intn(len(s))] }

func mapStatus(zendeskStatus string) string {
	switch zendeskStatus {
	case "new", "open":
		return "open"
	case "pending":
		return "in_progress"
	case "solved":
		return "resolved"
	case "closed":
		return "closed"
	default:
		return "open"
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: seed <org-slug>")
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

	var orgID string
	err = db.QueryRow(`SELECT id FROM organizations WHERE slug = $1`, slug).Scan(&orgID)
	if err == sql.ErrNoRows {
		log.Fatalf("org not found: %s", slug)
	}
	if err != nil {
		log.Fatalf("lookup org: %v", err)
	}
	log.Printf("seeding into org %s (id: %s)", slug, orgID)

	// Customers
	customerIDs := make([]string, 0, len(customerData))
	for _, c := range customerData {
		var id string
		err := db.QueryRow(
			`INSERT INTO customers (name, org_id) VALUES ($1, $2) RETURNING id`,
			c.name, orgID,
		).Scan(&id)
		if err != nil {
			log.Fatalf("insert customer %s: %v", c.name, err)
		}
		customerIDs = append(customerIDs, id)

		// Emails: first is verified, additional ones are not
		for i, email := range c.emails {
			_, err := db.Exec(
				`INSERT INTO customer_emails (customer_id, email, verified) VALUES ($1, $2, $3)`,
				id, email, i == 0,
			)
			if err != nil {
				log.Fatalf("insert customer email %s: %v", email, err)
			}
		}

		// Phones: first is verified, additional ones are not
		for i, phone := range c.phones {
			_, err := db.Exec(
				`INSERT INTO customer_phones (customer_id, phone, verified) VALUES ($1, $2, $3)`,
				id, phone, i == 0,
			)
			if err != nil {
				log.Fatalf("insert customer phone %s: %v", phone, err)
			}
		}
	}
	log.Printf("inserted %d customers", len(customerIDs))

	// Agents
	agentIDs := make([]string, 0, len(agentData))
	for _, a := range agentData {
		var id string
		err := db.QueryRow(
			`INSERT INTO agents (email, name, org_id) VALUES ($1, $2, $3) RETURNING id`,
			a.email, a.name, orgID,
		).Scan(&id)
		if err != nil {
			log.Fatalf("insert agent %s: %v", a.email, err)
		}
		agentIDs = append(agentIDs, id)
	}
	log.Printf("inserted %d agents", len(agentIDs))

	// Tickets
	ticketIDs := make([]string, 0, 50)
	for range 50 {
		reporterID := customerIDs[rand.Intn(len(customerIDs))]

		// ~30% of tickets are unassigned
		var assigneeID *string
		if rand.Float32() > 0.3 {
			id := agentIDs[rand.Intn(len(agentIDs))]
			assigneeID = &id
		}

		zendeskStatus := pick(zendeskStatuses)
		var id string
		err := db.QueryRow(
			`INSERT INTO tickets (title, description, status, priority, reporter_id, assignee_id, org_id, zendesk_status)
			 VALUES ($1, $2, $3::ticket_status, $4::ticket_priority, $5, $6, $7, $8::zendesk_status_category)
			 RETURNING id`,
			pick(titles),
			pick(descriptions),
			mapStatus(zendeskStatus),
			pick(priorities),
			reporterID,
			assigneeID,
			orgID,
			zendeskStatus,
		).Scan(&id)
		if err != nil {
			log.Fatalf("insert ticket: %v", err)
		}
		ticketIDs = append(ticketIDs, id)
	}
	log.Printf("inserted %d tickets", len(ticketIDs))

	// Comments: 3–5 per ticket, mix of customer and agent authors
	commentCount := 0
	for _, ticketID := range ticketIDs {
		n := 3 + rand.Intn(3)
		for range n {
			var customerAuthorID *string
			var agentAuthorID *string
			var role string

			if rand.Float32() < 0.4 {
				// Agent comment
				id := agentIDs[rand.Intn(len(agentIDs))]
				agentAuthorID = &id
				role = "agent"
			} else {
				// Customer comment
				id := customerIDs[rand.Intn(len(customerIDs))]
				customerAuthorID = &id
				role = "customer"
			}

			body := pick(customerCommentBodies)
			if role == "agent" {
				body = pick(agentCommentBodies)
			}

			_, err := db.Exec(
				`INSERT INTO comments (ticket_id, customer_author_id, agent_author_id, role, body)
				 VALUES ($1, $2, $3, $4::comment_role, $5)`,
				ticketID, customerAuthorID, agentAuthorID, role, body,
			)
			if err != nil {
				log.Fatalf("insert comment: %v", err)
			}
			commentCount++
		}
	}
	log.Printf("inserted %d comments", commentCount)

	// Place tickets into the default Kanban board columns by zendesk_status.
	var defaultBoardID string
	err = db.QueryRow(
		`SELECT id FROM boards WHERE org_id = $1 AND is_default = true`,
		orgID,
	).Scan(&defaultBoardID)
	if err == sql.ErrNoRows {
		log.Println("warn: no default Kanban board found — skipping ticket placement")
		return
	}
	if err != nil {
		log.Fatalf("query default board: %v", err)
	}

	result, err := db.Exec(`
		INSERT INTO kanban_board_tickets (board_id, column_id, ticket_id, position)
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
		log.Fatalf("place tickets in kanban: %v", err)
	}
	placed, _ := result.RowsAffected()
	log.Printf("placed %d tickets into default Kanban board", placed)
}
