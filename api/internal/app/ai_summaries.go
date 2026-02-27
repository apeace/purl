package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
)

// aiAnalysis is the JSON structure returned by the LLM for each ticket.
type aiAnalysis struct {
	Summary     string `json:"summary"`
	Temperature int    `json:"temperature"`
	Title       string `json:"title"`
}

// GenerateSummaries finds tickets where ai_summary IS NULL, ai_summary_stale = TRUE,
// or ai_title IS NULL, generates an AI analysis via Ollama for each, and updates the DB.
// Returns count of successfully updated tickets.
func GenerateSummaries(ctx context.Context, db *sql.DB, ollamaURL, ollamaModel string) (int, error) {
	client := newOllamaClient(ollamaURL, ollamaModel)

	if err := client.pullModel(ctx); err != nil {
		return 0, fmt.Errorf("pull model: %w", err)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, title, COALESCE(description, '')
		FROM tickets
		WHERE (ai_summary IS NULL OR ai_summary_stale = TRUE OR ai_title IS NULL)
		  AND ai_summary_error_count < 3
		ORDER BY updated_at DESC
		LIMIT 100`,
	)
	if err != nil {
		return 0, fmt.Errorf("query tickets: %w", err)
	}

	type ticketRow struct {
		id          string
		title       string
		description string
	}
	var tickets []ticketRow
	for rows.Next() {
		var t ticketRow
		if err := rows.Scan(&t.id, &t.title, &t.description); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan ticket: %w", err)
		}
		tickets = append(tickets, t)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate tickets: %w", err)
	}

	const workers = 2
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var updated atomic.Int32

	for _, t := range tickets {
		t := t
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			recordError := func(label string, err error) {
				msg := fmt.Sprintf("%s: %v", label, err)
				log.Printf("generate-summaries: ticket %s: %s", t.id, msg)
				if _, dbErr := db.ExecContext(ctx,
					`UPDATE tickets SET ai_summary_error_count = ai_summary_error_count + 1, ai_summary_last_error = $1 WHERE id = $2`,
					msg, t.id,
				); dbErr != nil {
					log.Printf("generate-summaries: ticket %s: increment error count: %v", t.id, dbErr)
				}
			}

			commentRows, err := db.QueryContext(ctx, `
				SELECT role, body
				FROM ticket_comments
				WHERE ticket_id = $1
				ORDER BY created_at ASC`,
				t.id,
			)
			if err != nil {
				recordError("query comments", err)
				return
			}

			var commentLines []string
			var firstCustomerBody string
			var scanErr bool
			for commentRows.Next() {
				var role, body string
				if err := commentRows.Scan(&role, &body); err != nil {
					commentRows.Close()
					recordError("scan comment", err)
					scanErr = true
					break
				}
				label := "[Customer]"
				if role == "agent" {
					label = "[Agent]"
				} else if firstCustomerBody == "" {
					firstCustomerBody = body
				}
				commentLines = append(commentLines, label+" "+body)
			}
			commentRows.Close()
			if scanErr {
				return
			}
			if err := commentRows.Err(); err != nil {
				recordError("iterate comments", err)
				return
			}

			// Fall back to ticket description if no customer comment was found.
			if firstCustomerBody == "" {
				firstCustomerBody = t.description
			}

			prompt := buildAnalysisPrompt(t.title, firstCustomerBody, commentLines)

			raw, err := client.chat(ctx, prompt, "json")
			if err != nil {
				recordError("chat", err)
				return
			}

			var analysis aiAnalysis
			if err := json.Unmarshal([]byte(raw), &analysis); err != nil {
				recordError("parse json response", fmt.Errorf("%w (raw: %q)", err, raw))
				return
			}
			if analysis.Title == "" || analysis.Summary == "" {
				recordError("incomplete json response", fmt.Errorf("title or summary missing (raw: %q)", raw))
				return
			}

			// ai_title uses COALESCE so it is only set on first generation and never overwritten.
			if _, err := db.ExecContext(ctx, `
				UPDATE tickets
				SET ai_title = COALESCE(ai_title, $1),
				    ai_summary = $2,
				    ai_temperature = $3,
				    ai_summary_stale = FALSE,
				    ai_summary_error_count = 0,
				    ai_summary_last_error = NULL
				WHERE id = $4`,
				analysis.Title, analysis.Summary, analysis.Temperature, t.id,
			); err != nil {
				log.Printf("generate-summaries: ticket %s: update: %v", t.id, err)
				return
			}
			updated.Add(1)
		}()
	}
	wg.Wait()

	return int(updated.Load()), nil
}

func buildAnalysisPrompt(title, firstInquiry string, commentLines []string) string {
	var b strings.Builder
	b.WriteString("Analyze this customer support ticket and respond with a JSON object with exactly these three fields:\n")
	b.WriteString("- \"title\": A 4-5 word title based only on the customer's first message (e.g. \"Login issue with SSO\")\n")
	b.WriteString("- \"summary\": 1-2 terse sentences about the current state of the ticket — issue and status only, no intro phrases like \"The customer\" or \"This ticket\"\n")
	b.WriteString("- \"temperature\": An integer 1-10 for how urgent or frustrated the customer is (1 = patient and calm, 10 = furious or threatening to cancel)\n")
	b.WriteString("All three fields are required. Never leave title or summary empty.\n")
	b.WriteString("Example response: {\"title\": \"Login issue with SSO\", \"summary\": \"SSO login fails with a 403 error. Agent requested HAR file.\", \"temperature\": 4}\n")
	b.WriteString("Output only the JSON object, nothing else.\n\n")
	b.WriteString("Ticket subject: ")
	b.WriteString(title)
	b.WriteString("\nCustomer's first message: ")
	b.WriteString(firstInquiry)
	if len(commentLines) > 0 {
		b.WriteString("\n\nFull conversation:\n")
		b.WriteString(strings.Join(commentLines, "\n"))
	}
	return b.String()
}
