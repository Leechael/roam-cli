package roam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"roam-cli/internal/config"
)

const (
	// maxActionsPerBatch is the maximum number of actions sent in a single
	// batch-actions request. Roam recommends 50-100; we use 50 to stay safe.
	maxActionsPerBatch = 50

	// delayBetweenBatches is the pause between consecutive chunk requests
	// to stay within ~2-3 req/s rate guidance.
	delayBetweenBatches = 400 * time.Millisecond

	// maxRetries is the number of retries on HTTP 429.
	maxRetries = 5
)

type Client struct {
	httpClient *http.Client
	baseGraph  string
	token      string
}

type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("roam api error: status=%d body=%s", e.Status, e.Body)
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second},
		baseGraph:  fmt.Sprintf("%s/%s", cfg.BaseURL, cfg.Graph),
		token:      cfg.Token,
	}
}

func (c *Client) Q(query string, args []string) (any, error) {
	payload := map[string]any{
		"query": query,
		"args":  args,
	}
	resp, err := c.postJSON("/q", payload)
	if err != nil {
		return nil, err
	}
	result, ok := resp["result"]
	if !ok {
		return nil, nil
	}
	return result, nil
}

func (c *Client) Write(payload map[string]any) (map[string]any, error) {
	return c.postJSON("/write", payload)
}

// BatchActions sends actions in chunks of maxActionsPerBatch, pausing
// between chunks to respect Roam API rate limits. Each chunk is retried
// with exponential backoff on HTTP 429.
func (c *Client) BatchActions(actions []map[string]any) (map[string]any, error) {
	if len(actions) <= maxActionsPerBatch {
		return c.batchOnce(actions)
	}

	var lastResp map[string]any
	for i := 0; i < len(actions); i += maxActionsPerBatch {
		end := i + maxActionsPerBatch
		if end > len(actions) {
			end = len(actions)
		}
		resp, err := c.batchOnce(actions[i:end])
		if err != nil {
			return nil, fmt.Errorf("batch chunk [%d:%d] failed: %w", i, end, err)
		}
		lastResp = resp

		if end < len(actions) {
			time.Sleep(delayBetweenBatches)
		}
	}
	return lastResp, nil
}

func (c *Client) batchOnce(actions []map[string]any) (map[string]any, error) {
	payload := map[string]any{
		"action":  "batch-actions",
		"actions": actions,
	}
	return c.Write(payload)
}

func (c *Client) postJSON(path string, payload map[string]any) (map[string]any, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	for attempt := 0; ; attempt++ {
		req, err := http.NewRequest(http.MethodPost, c.baseGraph+path, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		req.Header.Set("X-Authorization", fmt.Sprintf("Bearer %s", c.token))

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		respBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests && attempt < maxRetries {
			wait := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			time.Sleep(wait)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, &APIError{Status: resp.StatusCode, Body: string(respBytes)}
		}
		if len(respBytes) == 0 {
			return map[string]any{"ok": true}, nil
		}

		var data map[string]any
		if err := json.Unmarshal(respBytes, &data); err != nil {
			return map[string]any{"raw": string(respBytes)}, nil
		}
		return data, nil
	}
}
