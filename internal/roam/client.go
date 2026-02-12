package roam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"roam-cli/internal/config"
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

func (c *Client) BatchActions(actions []map[string]any) (map[string]any, error) {
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
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
