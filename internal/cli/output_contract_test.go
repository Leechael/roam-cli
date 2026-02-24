package cli

import (
	"strings"
	"testing"
)

func TestOutputModeMutualExclusion(t *testing.T) {
	tests := []struct {
		name    string
		cmd     func() error
		errText string
	}{
		{
			name: "status json and plain conflict",
			cmd: func() error {
				opts = globalOptions{}
				c := newStatusCmd()
				c.SetArgs([]string{"--json", "--plain"})
				return c.Execute()
			},
			errText: "--json and --plain cannot be used together",
		},
		{
			name: "get json and plain conflict",
			cmd: func() error {
				opts = globalOptions{}
				c := newGetCmd()
				c.SetArgs([]string{"AnyTitle", "--json", "--plain"})
				return c.Execute()
			},
			errText: "--json and --plain cannot be used together",
		},
		{
			name: "search json and plain conflict",
			cmd: func() error {
				opts = globalOptions{}
				c := newSearchCmd()
				c.SetArgs([]string{"hello", "--json", "--plain"})
				return c.Execute()
			},
			errText: "--json and --plain cannot be used together",
		},
		{
			name: "q json and plain conflict",
			cmd: func() error {
				opts = globalOptions{}
				c := newQCmd()
				c.SetArgs([]string{"[:find ?e :where [?e :node/title]]", "--json", "--plain"})
				return c.Execute()
			},
			errText: "--json and --plain cannot be used together",
		},
		{
			name: "save json and plain conflict",
			cmd: func() error {
				opts = globalOptions{}
				c := newSaveCmd()
				c.SetArgs([]string{"--title", "T", "--json", "--plain"})
				return c.Execute()
			},
			errText: "--json and --plain cannot be used together",
		},
		{
			name: "journal json and plain conflict",
			cmd: func() error {
				opts = globalOptions{}
				c := newJournalCmd()
				c.SetArgs([]string{"--json", "--plain"})
				return c.Execute()
			},
			errText: "--json and --plain cannot be used together",
		},
		{
			name: "block create json and plain conflict",
			cmd: func() error {
				opts = globalOptions{}
				c := newBlockCreateCmd()
				c.SetArgs([]string{"--json", "--plain", "--parent", "p", "--text", "x"})
				return c.Execute()
			},
			errText: "--json and --plain cannot be used together",
		},
		{
			name: "block move json and plain conflict",
			cmd: func() error {
				opts = globalOptions{}
				c := newBlockMoveCmd()
				c.SetArgs([]string{"--json", "--plain", "--uid", "u", "--parent", "p"})
				return c.Execute()
			},
			errText: "--json and --plain cannot be used together",
		},
		{
			name: "batch run json and plain conflict",
			cmd: func() error {
				opts = globalOptions{}
				root := newBatchCmd()
				root.SetArgs([]string{"run", "--json", "--plain", "--stdin"})
				return root.Execute()
			},
			errText: "--json and --plain cannot be used together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd()
			if err == nil {
				t.Fatalf("expected error")
			}
			if !strings.Contains(err.Error(), tt.errText) {
				t.Fatalf("expected error to contain %q, got %q", tt.errText, err.Error())
			}
		})
	}
}

func TestQJQRequiresJSON(t *testing.T) {
	opts = globalOptions{}
	cmd := newQCmd()
	cmd.SetArgs([]string{"[:find ?e :where [?e :node/title]]", "--jq", "."})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected non-nil error when --jq is used without --json")
	}
	if !strings.Contains(err.Error(), "--jq requires --json") {
		t.Fatalf("unexpected error: %v", err)
	}
}
