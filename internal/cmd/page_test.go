package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setTestAPIEnv(t *testing.T, srv *httptest.Server) {
	t.Helper()
	t.Setenv("ROAM_API_TOKEN", "test-token")
	t.Setenv("ROAM_API_GRAPH", "test-graph")
	t.Setenv("ROAM_API_BASE_URL", srv.URL+"/api/graph")
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	out := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		out <- buf.String()
	}()

	fn()
	_ = w.Close()
	return <-out
}

func readJSONBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	var body map[string]any
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	return body
}

func TestSaveReplace_ClearsExistingPage(t *testing.T) {
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		body := readJSONBody(t, r)
		switch {
		case r.URL.Path == "/api/graph/test-graph/q" && strings.Contains(body["query"].(string), `:find ?uid`) && strings.Contains(body["query"].(string), `:node/title "Demo Page"`):
			_, _ = w.Write([]byte(`{"result":[["page-uid"]]}`))
		case r.URL.Path == "/api/graph/test-graph/q" && strings.Contains(body["query"].(string), `{:block/children ...}`):
			page := map[string]any{
				":block/uid": "page-uid",
				":block/children": []any{
					map[string]any{":block/uid": "child-uid", ":block/order": 0},
				},
			}
			resp, _ := json.Marshal(map[string]any{"result": []any{[]any{page}}})
			_, _ = w.Write(resp)
		case r.URL.Path == "/api/graph/test-graph/write":
			actions, _ := body["actions"].([]any)
			if len(actions) == 0 {
				t.Fatalf("expected batch actions in write request")
			}
			if call == 3 {
				action := actions[0].(map[string]any)
				if action["action"] != "delete-block" {
					t.Fatalf("expected delete-block action, got %#v", action)
				}
				block := action["block"].(map[string]any)
				if block["uid"] != "child-uid" {
					t.Fatalf("expected child-uid, got %#v", block["uid"])
				}
			} else {
				action := actions[0].(map[string]any)
				if action["action"] != "create-block" {
					t.Fatalf("expected create-block action, got %#v", action)
				}
				block := action["block"].(map[string]any)
				if block["string"] != "new item" {
					t.Fatalf("expected new item, got %#v", block["string"])
				}
			}
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			t.Fatalf("unexpected request: %s %s body=%v", r.Method, r.URL.Path, body)
		}
	}))
	defer srv.Close()
	setTestAPIEnv(t, srv)

	file := filepath.Join(t.TempDir(), "note.md")
	if err := os.WriteFile(file, []byte("- new item\n"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	root := newRootCmd()
	root.SetArgs([]string{"save", "--title", "Demo Page", "--replace", "--file", file})
	out := captureStdout(t, func() {
		if err := root.Execute(); err != nil {
			t.Fatalf("unexpected error")
		}
	})
	if !strings.Contains(out, "replaced page") {
		t.Fatalf("expected replaced page output, got: %s", out)
	}
	if call != 4 {
		t.Fatalf("expected 4 requests, got %d", call)
	}
}

func TestPageClearAndDelete(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "clear", args: []string{"page", "clear", "Demo Page", "--plain"}, want: "page-uid"},
		{name: "delete", args: []string{"page", "delete", "Demo Page", "--plain"}, want: "page-uid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				call++
				body := readJSONBody(t, r)
				switch {
				case tt.name == "clear" && r.URL.Path == "/api/graph/test-graph/q" && strings.Contains(body["query"].(string), `{:block/children ...}`):
					page := map[string]any{
						":block/uid": "page-uid",
						":block/children": []any{
							map[string]any{":block/uid": "child-uid", ":block/order": 0},
						},
					}
					resp, _ := json.Marshal(map[string]any{"result": []any{[]any{page}}})
					_, _ = w.Write(resp)
				case tt.name == "delete" && r.URL.Path == "/api/graph/test-graph/q":
					_, _ = w.Write([]byte(`{"result":[["page-uid"]]}`))
				case r.URL.Path == "/api/graph/test-graph/write":
					actions, _ := body["actions"].([]any)
					if tt.name == "clear" {
						if len(actions) != 1 {
							t.Fatalf("expected 1 delete action, got %d", len(actions))
						}
						action := actions[0].(map[string]any)
						if action["action"] != "delete-block" {
							t.Fatalf("expected delete-block action, got %#v", action)
						}
						block := action["block"].(map[string]any)
						if block["uid"] != "child-uid" {
							t.Fatalf("expected child-uid, got %#v", block["uid"])
						}
					} else {
						if body["action"] != "delete-block" {
							t.Fatalf("expected delete-block request, got %#v", body["action"])
						}
						block := body["block"].(map[string]any)
						if block["uid"] != "page-uid" {
							t.Fatalf("expected page-uid, got %#v", block["uid"])
						}
					}
					_, _ = w.Write([]byte(`{"ok":true}`))
				default:
					t.Fatalf("unexpected request: %s %s body=%v", r.Method, r.URL.Path, body)
				}
			}))
			defer srv.Close()
			setTestAPIEnv(t, srv)

			root := newRootCmd()
			root.SetArgs(tt.args)
			out := captureStdout(t, func() {
				if err := root.Execute(); err != nil {
					t.Fatalf("unexpected error")
				}
			})
			if strings.TrimSpace(out) != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, strings.TrimSpace(out))
			}
		})
	}
}
