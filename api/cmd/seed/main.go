package main

import (
	"database/sql"
	"log"
	"math/rand"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var users = []struct{ email, name string }{
	{"alice@example.com", "Alice Chen"},
	{"bob@example.com", "Bob Harris"},
	{"carol@example.com", "Carol Singh"},
	{"dan@example.com", "Dan Okafor"},
	{"eve@example.com", "Eve Nakamura"},
	{"frank@example.com", "Frank Torres"},
	{"grace@example.com", "Grace Müller"},
	{"henry@example.com", "Henry Park"},
	{"isla@example.com", "Isla Rivera"},
	{"james@example.com", "James Kowalski"},
}

var titles = []string{
	"Login page throws 500 on bad password",
	"Dashboard fails to load for new accounts",
	"Email notifications not sent after ticket close",
	"Search returns stale results",
	"Attachment upload silently fails over 5 MB",
	"Password reset link expires too quickly",
	"Dark mode flickers on page load",
	"CSV export includes deleted records",
	"Sorting by priority broken on mobile",
	"Session expires mid-form and loses data",
	"API rate limit not documented",
	"Duplicate tickets can be created via rapid clicks",
	"Pagination skips a page at record boundary",
	"Assignee field resets on ticket reopen",
	"Webhook payload missing ticket URL",
	"Add bulk status update action",
	"Support markdown in ticket descriptions",
	"Allow filtering by multiple assignees",
	"Add keyboard shortcut to create ticket",
	"Show time-in-status on ticket detail",
	"Integrate Slack notifications",
	"Add due date field to tickets",
	"Display reporter avatar in list view",
	"Allow ticket templates",
	"Export tickets to PDF",
	"Improve error messages on form validation",
	"Add two-factor authentication",
	"Remember table column widths per user",
	"Mobile: tap target too small on ticket row",
	"Tooltip overlaps dropdown on small screens",
}

var descriptions = []string{
	"Steps to reproduce are attached. Happens consistently in production.",
	"Reported by multiple users. Likely related to the recent deploy.",
	"Intermittent — hard to reproduce locally but reliable in staging.",
	"Blocked by this issue. Needs fix before the next release.",
	"Low priority but affects a notable number of users per analytics.",
	"Quick win — the fix should be straightforward.",
	"Needs design review before implementation.",
	"Regression introduced in v2.3. Was working before.",
}

// Weighted toward open/in_progress to reflect a realistic backlog.
var statuses = []string{
	"open", "open", "open",
	"in_progress", "in_progress",
	"resolved",
	"closed",
}

// Weighted toward medium.
var priorities = []string{
	"low",
	"medium", "medium", "medium",
	"high",
	"urgent",
}

var commentBodies = []string{
	"I can reproduce this consistently on Chrome 122.",
	"Tried the workaround mentioned above — no luck.",
	"This is blocking us. Bumping priority.",
	"Fix looks good in staging. Ready for review.",
	"Rolled back the change. Will investigate further.",
	"Confirmed fixed in the latest build.",
	"Added a failing test case to the branch.",
	"Waiting on design feedback before proceeding.",
	"Is there a workaround we can use in the meantime?",
	"Related to #42 — might share the same root cause.",
	"I have a patch ready, will open a PR shortly.",
	"This also affects the mobile app.",
	"Closing as duplicate — tracked in the other ticket.",
	"Left a comment in the code explaining the fix.",
	"Spoke with the reporter — they can live with this for now.",
}

func pick(s []string) string { return s[rand.Intn(len(s))] }

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://pipeline:pipeline@localhost:5432/pipeline?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose set dialect: %v", err)
	}
	// Uses the OS filesystem so this must be run from the api/ directory.
	// Reset drops all tables, Up recreates them — ensures a clean slate even
	// if the schema has changed since the migration was last applied.
	if err := goose.Reset(db, "migrations"); err != nil {
		log.Fatalf("goose reset: %v", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		log.Fatalf("goose up: %v", err)
	}
	log.Println("reset and migrated")

	// Users
	userIDs := make([]string, 0, len(users))
	for _, u := range users {
		var id string
		err := db.QueryRow(
			`INSERT INTO users (email, name) VALUES ($1, $2) RETURNING id`,
			u.email, u.name,
		).Scan(&id)
		if err != nil {
			log.Fatalf("insert user %s: %v", u.email, err)
		}
		userIDs = append(userIDs, id)
	}
	log.Printf("inserted %d users", len(userIDs))

	// Tickets
	ticketIDs := make([]string, 0, 50)
	for range 50 {
		reporterID := userIDs[rand.Intn(len(userIDs))]

		// ~30% of tickets are unassigned
		var assigneeID *string
		if rand.Float32() > 0.3 {
			id := userIDs[rand.Intn(len(userIDs))]
			assigneeID = &id
		}

		var id string
		err := db.QueryRow(
			`INSERT INTO tickets (title, description, status, priority, reporter_id, assignee_id)
			 VALUES ($1, $2, $3::ticket_status, $4::ticket_priority, $5, $6)
			 RETURNING id`,
			pick(titles),
			pick(descriptions),
			pick(statuses),
			pick(priorities),
			reporterID,
			assigneeID,
		).Scan(&id)
		if err != nil {
			log.Fatalf("insert ticket: %v", err)
		}
		ticketIDs = append(ticketIDs, id)
	}
	log.Printf("inserted %d tickets", len(ticketIDs))

	// Comments: 3–5 per ticket
	commentCount := 0
	for _, ticketID := range ticketIDs {
		n := 3 + rand.Intn(3)
		for range n {
			authorID := userIDs[rand.Intn(len(userIDs))]
			_, err := db.Exec(
				`INSERT INTO comments (ticket_id, author_id, body) VALUES ($1, $2, $3)`,
				ticketID, authorID, pick(commentBodies),
			)
			if err != nil {
				log.Fatalf("insert comment: %v", err)
			}
			commentCount++
		}
	}
	log.Printf("inserted %d comments", commentCount)
}
