package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
)

// GenerateSummaries finds tickets where ai_summary IS NULL or ai_summary_stale = TRUE,
// generates a summary via Ollama for each, and updates the DB.
// Returns count of successfully updated tickets.
func GenerateSummaries(ctx context.Context, db *sql.DB, ollamaURL, ollamaModel string) (int, error) {
	client := newOllamaClient(ollamaURL, ollamaModel)

	if err := client.pullModel(ctx); err != nil {
		return 0, fmt.Errorf("pull model: %w", err)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT id, title, COALESCE(description, '')
		FROM tickets
		WHERE (ai_summary IS NULL OR ai_summary_stale = TRUE)
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

			prompt := buildSummaryPrompt(t.title, t.description, commentLines)

			summary, err := client.chat(ctx, prompt)
			if err != nil {
				recordError("chat", err)
				return
			}

			if _, err := db.ExecContext(ctx,
				`UPDATE tickets SET ai_summary = $1, ai_summary_stale = FALSE, ai_summary_error_count = 0, ai_summary_last_error = NULL WHERE id = $2`,
				summary, t.id,
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

func buildSummaryPrompt(title, description string, commentLines []string) string {
	var b strings.Builder
	b.WriteString("Summarize this customer support ticket in 1-2 short, direct sentences. State only the issue and current status — no intro phrase like \"The customer\" or \"This ticket\". Output only the summary, nothing else.\n\n")
	b.WriteString("Title: ")
	b.WriteString(title)
	b.WriteString("\nDescription: ")
	b.WriteString(description)
	if len(commentLines) > 0 {
		b.WriteString("\n\nComments:\n")
		b.WriteString(strings.Join(commentLines, "\n"))
	}
	return b.String()
}
