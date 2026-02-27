package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ollamaClient struct {
	baseURL string
	model   string
}

func newOllamaClient(baseURL, model string) *ollamaClient {
	return &ollamaClient{baseURL: baseURL, model: model}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Format   string        `json:"format,omitempty"`
	Messages []chatMessage `json:"messages"`
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
}

type chatResponse struct {
	Message chatMessage `json:"message"`
}

// pullModel ensures the model is present locally. It is idempotent — if the
// model is already downloaded, Ollama responds quickly. The first run may
// download several hundred MB, so a 30-minute timeout is used.
func (c *ollamaClient) pullModel(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	body, err := json.Marshal(map[string]any{"name": c.model, "stream": false})
	if err != nil {
		return fmt.Errorf("marshal pull request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/pull", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build pull request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama pull: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama pull returned %s: %s", resp.Status, b)
	}
	return nil
}

// chat sends a single-turn prompt to the model and returns the assistant reply.
// Uses a 5-minute timeout to accommodate slow local inference.
// Pass format="json" to instruct Ollama to return valid JSON.
func (c *ollamaClient) chat(ctx context.Context, prompt, format string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	body, err := json.Marshal(chatRequest{
		Format:   format,
		Messages: []chatMessage{{Role: "user", Content: prompt}},
		Model:    c.model,
		Stream:   false,
	})
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama chat: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read chat response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama chat returned %s: %s", resp.Status, respBody)
	}

	var result chatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse chat response: %w", err)
	}
	return result.Message.Content, nil
}
