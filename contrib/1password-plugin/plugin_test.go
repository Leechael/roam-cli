package main

import (
	"testing"

	"github.com/1Password/shell-plugins/sdk"
)

func TestPluginName(t *testing.T) {
	p := New()
	if p.Name != "roamresearch" {
		t.Errorf("expected plugin name %q, got %q", "roamresearch", p.Name)
	}
}

func TestPluginHasCredentials(t *testing.T) {
	p := New()
	if len(p.Credentials) == 0 {
		t.Fatal("plugin has no credential types")
	}
}

func TestPluginHasExecutables(t *testing.T) {
	p := New()
	if len(p.Executables) == 0 {
		t.Fatal("plugin has no executables")
	}
}

func TestNeedsAuthReturnsFalseForNoArgs(t *testing.T) {
	exec := RoamCLI()
	if exec.NeedsAuth == nil {
		t.Fatal("NeedsAuth is nil")
	}

	input := sdk.NeedsAuthenticationInput{
		CommandArgs: []string{},
	}
	if exec.NeedsAuth(input) {
		t.Error("expected NeedsAuth to be false for no args")
	}
}

func TestNeedsAuthReturnsFalseForHelp(t *testing.T) {
	exec := RoamCLI()

	tests := []struct {
		name string
		args []string
	}{
		{"--help", []string{"--help"}},
		{"help subcommand", []string{"help"}},
		{"--version", []string{"--version"}},
		{"completion command", []string{"completion", "zsh"}},
		{"cobra shell completion", []string{"__complete", "get", ""}},
		{"cobra shell completion no descriptions", []string{"__completeNoDesc", "get", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := sdk.NeedsAuthenticationInput{
				CommandArgs: tt.args,
			}
			if exec.NeedsAuth(input) {
				t.Errorf("expected NeedsAuth to be false for args %v", tt.args)
			}
		})
	}
}

func TestNeedsAuthReturnsTrueForStatus(t *testing.T) {
	exec := RoamCLI()

	input := sdk.NeedsAuthenticationInput{
		CommandArgs: []string{"status"},
	}
	if !exec.NeedsAuth(input) {
		t.Error("expected NeedsAuth to be true for 'status'")
	}
}

func TestNeedsAuthReturnsTrueForGetWithArgs(t *testing.T) {
	exec := RoamCLI()

	input := sdk.NeedsAuthenticationInput{
		CommandArgs: []string{"get", "--today"},
	}
	if !exec.NeedsAuth(input) {
		t.Error("expected NeedsAuth to be true for 'get --today'")
	}
}

func TestSchemaDeepValidationHasNoErrors(t *testing.T) {
	for _, report := range New().DeepValidate() {
		if !report.HasErrors() {
			continue
		}
		for _, check := range report.Checks {
			if !check.Assertion {
				t.Errorf("%s: %s", report.Heading, check.Description)
			}
		}
	}
}
