package client

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Leechael/roamresearch-skills/internal/config"
	"github.com/Leechael/roamresearch-skills/internal/model"
)

func testServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	srv := httptest.NewServer(handler)
	cfg := &config.Config{
		Token:          "test-token",
		Graph:          "testgraph",
		BaseURL:        srv.URL + "/api/graph",
		TimeoutSeconds: 5,
	}
	return srv, NewClient(cfg)
}

func TestQ_Success(t *testing.T) {
	srv, c := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/graph/testgraph/q" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		auth := r.Header.Get("X-Authorization")
		if auth != "Bearer test-token" {
			t.Fatalf("unexpected auth header: %s", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": []any{[]any{"Hello"}},
		})
	})
	defer srv.Close()

	result, err := c.Q("[:find ?t :where [?e :node/title ?t]]", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rows, ok := result.([]any)
	if !ok || len(rows) == 0 {
		t.Fatalf("expected non-empty result, got: %v", result)
	}
}

func TestQ_AuthError(t *testing.T) {
	srv, c := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	})
	defer srv.Close()

	_, err := c.Q("[:find ?t :where [?e :node/title ?t]]", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *model.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got: %T", err)
	}
	if apiErr.Status != 401 {
		t.Fatalf("expected status 401, got %d", apiErr.Status)
	}
}

func TestQ_NotFound(t *testing.T) {
	srv, c := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	})
	defer srv.Close()

	_, err := c.Q("[:find ?t :where [?e :node/title ?t]]", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *model.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got: %T", err)
	}
	if apiErr.Status != 404 {
		t.Fatalf("expected status 404, got %d", apiErr.Status)
	}
}

func TestWrite_Success(t *testing.T) {
	srv, c := testServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/graph/testgraph/write" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	defer srv.Close()

	resp, err := c.Write(map[string]any{
		"action": "create-block",
		"block":  map[string]any{"string": "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["ok"] != true {
		t.Fatalf("expected ok=true, got: %v", resp)
	}
}

func TestBatchActions_SingleChunk(t *testing.T) {
	callCount := 0
	srv, c := testServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	defer srv.Close()

	actions := make([]map[string]any, 5)
	for i := range actions {
		actions[i] = map[string]any{"action": "update-block", "block": map[string]any{"uid": "u"}}
	}
	_, err := c.BatchActions(actions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 API call, got %d", callCount)
	}
}

func TestQ_EmptyBody(t *testing.T) {
	srv, c := testServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	result, err := c.Q("[:find ?t]", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty body returns {"ok": true}, result key missing → nil
	if result != nil {
		t.Fatalf("expected nil result for empty body, got: %v", result)
	}
}
