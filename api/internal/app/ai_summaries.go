package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
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
		WHERE ai_summary IS NULL OR ai_summary_stale = TRUE
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

	updated := 0
	for _, t := range tickets {
		commentRows, err := db.QueryContext(ctx, `
			SELECT role, body
			FROM ticket_comments
			WHERE ticket_id = $1
			ORDER BY created_at ASC`,
			t.id,
		)
		if err != nil {
			log.Printf("generate-summaries: ticket %s: query comments: %v", t.id, err)
			continue
		}

		var commentLines []string
		for commentRows.Next() {
			var role, body string
			if err := commentRows.Scan(&role, &body); err != nil {
				commentRows.Close()
				log.Printf("generate-summaries: ticket %s: scan comment: %v", t.id, err)
				break
			}
			label := "[Customer]"
			if role == "agent" {
				label = "[Agent]"
			}
			commentLines = append(commentLines, label+" "+body)
		}
		commentRows.Close()
		if err := commentRows.Err(); err != nil {
			log.Printf("generate-summaries: ticket %s: iterate comments: %v", t.id, err)
			continue
		}

		prompt := buildSummaryPrompt(t.title, t.description, commentLines)

		summary, err := client.chat(ctx, prompt)
		if err != nil {
			log.Printf("generate-summaries: ticket %s: chat: %v", t.id, err)
			continue
		}

		if _, err := db.ExecContext(ctx,
			`UPDATE tickets SET ai_summary = $1, ai_summary_stale = FALSE WHERE id = $2`,
			summary, t.id,
		); err != nil {
			log.Printf("generate-summaries: ticket %s: update: %v", t.id, err)
			continue
		}
		updated++
	}

	return updated, nil
}

func buildSummaryPrompt(title, description string, commentLines []string) string {
	var b strings.Builder
	b.WriteString("Summarize this customer support ticket in 1-2 sentences. Focus on the main issue and current status.\n\n")
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
