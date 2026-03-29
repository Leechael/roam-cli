package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		json    bool
		plain   bool
		jq      string
		wantErr string
	}{
		{name: "json only", json: true},
		{name: "plain only", plain: true},
		{name: "default"},
		{name: "json+jq", json: true, jq: "."},
		{name: "json+plain conflict", json: true, plain: true, wantErr: "--json and --plain cannot be used together"},
		{name: "jq without json", jq: ".", wantErr: "--jq requires --json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New(tt.json, tt.plain, tt.jq)
			err := f.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

func TestPrintJSON(t *testing.T) {
	f := New(true, false, "")
	var buf bytes.Buffer
	err := f.Print(&buf, map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"key": "value"`) {
		t.Fatalf("expected pretty JSON, got: %s", out)
	}
}

func TestPrintJQ(t *testing.T) {
	f := New(true, false, ".key")
	var buf bytes.Buffer
	err := f.Print(&buf, map[string]any{"key": "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out != `"hello"` {
		t.Fatalf("expected \"hello\", got: %s", out)
	}
}

func TestPrintDefault_NoOutput(t *testing.T) {
	f := New(false, false, "")
	var buf bytes.Buffer
	err := f.Print(&buf, map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output in default mode, got: %s", buf.String())
	}
}

func TestPrintPlain_NoOutput(t *testing.T) {
	f := New(false, true, "")
	var buf bytes.Buffer
	err := f.Print(&buf, map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output in plain mode, got: %s", buf.String())
	}
}

func TestIsJSON(t *testing.T) {
	f := New(true, false, "")
	if !f.IsJSON() {
		t.Fatal("expected IsJSON true")
	}
	if f.IsPlain() {
		t.Fatal("expected IsPlain false")
	}
}
