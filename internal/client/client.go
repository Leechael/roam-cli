package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/Leechael/roam-cli/internal/config"
	"github.com/Leechael/roam-cli/internal/model"
)

const (
	// maxActionsPerBatch is the maximum number of actions sent in a single
	// batch-actions request. Roam recommends 50-100; we use 30 to leave
	// headroom for other calls in the same session.
	maxActionsPerBatch = 30

	// delayBetweenBatches is the pause between consecutive chunk requests.
	// Roam rate limit is ~100 req/min; 2s keeps us well under 30 req/min
	// even for large batches, leaving budget for surrounding calls.
	delayBetweenBatches = 2 * time.Second

	// maxRetries is the number of retries on HTTP 429.
	// Backoff: 2, 4, 8, 16, 32, 64s = 126s total, enough to wait out
	// a full 60s rate-limit window.
	maxRetries = 6

	// retryBaseDelay is the base for exponential backoff on 429.
	retryBaseDelay = 2 * time.Second
)

type Client struct {
	httpClient *http.Client
	baseGraph  string
	token      string
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
			wait := time.Duration(float64(retryBaseDelay) * math.Pow(2, float64(attempt)))
			time.Sleep(wait)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, &model.APIError{Status: resp.StatusCode, Body: string(respBytes)}
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
