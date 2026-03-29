package cmd

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files")

func TestRootHelp_Golden(t *testing.T) {
	root := newRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"--help"})
	_ = root.Execute()

	goldenPath := filepath.Join("testdata", "root-help.golden")
	assertGolden(t, goldenPath, buf.String())
}

func TestHelpExitCodes_Golden(t *testing.T) {
	root := newRootCmd()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"help", "exit-codes"})
	_ = root.Execute()

	goldenPath := filepath.Join("testdata", "help-exit-codes.golden")
	assertGolden(t, goldenPath, buf.String())
}

func TestHelpUnknownTopic(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"help", "nonexistent-topic"})
	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for unknown topic")
	}
	errMsg := err.Error()
	if !bytes.Contains([]byte(errMsg), []byte("unknown help topic")) {
		t.Fatalf("expected 'unknown help topic' in error, got: %s", errMsg)
	}
	if !bytes.Contains([]byte(errMsg), []byte("exit-codes")) {
		t.Fatalf("expected available topics in error, got: %s", errMsg)
	}
}

func assertGolden(t *testing.T, goldenPath string, actual string) {
	t.Helper()
	if *update {
		if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
			t.Fatalf("failed to update golden file: %v", err)
		}
		return
	}
	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v\nRun with -update to create it", goldenPath, err)
	}
	if actual != string(expected) {
		t.Fatalf("output does not match golden file %s\n--- want ---\n%s\n--- got ---\n%s", goldenPath, string(expected), actual)
	}
}
