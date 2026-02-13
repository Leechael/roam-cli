package cli

import "testing"

func TestStatusJSONFailureReturnsError(t *testing.T) {
	// Ensure no credentials are available from env or flags.
	t.Setenv("ROAM_API_TOKEN", "")
	t.Setenv("ROAM_API_GRAPH", "")
	t.Setenv("ROAM_API_BASE_URL", "")
	t.Setenv("ROAM_TIMEOUT_SECONDS", "")
	opts = globalOptions{}

	cmd := newStatusCmd()
	cmd.SetArgs([]string{"--json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected non-nil error when status check fails with --json")
	}
}

func TestStatusJQRequiresJSON(t *testing.T) {
	opts = globalOptions{}

	cmd := newStatusCmd()
	cmd.SetArgs([]string{"--jq", ".ok"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected non-nil error when --jq is used without --json")
	}
}
