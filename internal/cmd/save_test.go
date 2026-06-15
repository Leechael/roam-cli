package cmd

import (
	"bytes"
	"testing"
)

func TestValidateSaveTarget(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		parentUID string
		dailyPage string
		today     bool
		wantErr   bool
	}{
		{name: "page mode ok", title: "Page", wantErr: false},
		{name: "parent mode ok", parentUID: "p1", wantErr: false},
		{name: "daily page ok", dailyPage: "2026-03-22", wantErr: false},
		{name: "today ok", today: true, wantErr: false},
		{name: "missing all", wantErr: true},
		{name: "title + parent", title: "Page", parentUID: "p1", wantErr: true},
		{name: "title + daily", title: "Page", dailyPage: "2026-03-22", wantErr: true},
		{name: "title + today", title: "Page", today: true, wantErr: true},
		{name: "parent + daily", parentUID: "p1", dailyPage: "2026-03-22", wantErr: true},
		{name: "parent + today", parentUID: "p1", today: true, wantErr: true},
		{name: "daily + today", dailyPage: "2026-03-22", today: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSaveTarget(tt.title, tt.parentUID, tt.dailyPage, tt.today)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSaveHelpMentionsHelpTopics(t *testing.T) {
	root := newRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"save", "--help"})
	_ = root.Execute()

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("roam-cli help writing-guide")) {
		t.Fatalf("expected save help to mention writing-guide, got: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("roam-cli help write-examples")) {
		t.Fatalf("expected save help to mention write-examples, got: %s", out)
	}
}
