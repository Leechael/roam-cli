package main

import (
	"context"
	"testing"

	"github.com/1Password/shell-plugins/sdk"
)

func TestCredentialTypeName(t *testing.T) {
	ct := APIToken()
	if string(ct.Name) != "API Token" {
		t.Errorf("expected %q, got %q", "API Token", ct.Name)
	}
}

func TestCredentialFields(t *testing.T) {
	ct := APIToken()

	if len(ct.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(ct.Fields))
	}

	tests := []struct {
		name     sdk.FieldName
		secret   bool
		optional bool
	}{
		{fieldToken, true, false},
		{fieldGraph, false, false},
		{fieldAPIURL, false, true},
		{fieldTimeoutSeconds, false, true},
	}

	for _, tt := range tests {
		found := false
		for _, f := range ct.Fields {
			if f.Name == tt.name {
				found = true
				if f.Secret != tt.secret {
					t.Errorf("field %q: secret = %v, want %v", tt.name, f.Secret, tt.secret)
				}
				if f.Optional != tt.optional {
					t.Errorf("field %q: optional = %v, want %v", tt.name, f.Optional, tt.optional)
				}
				break
			}
		}
		if !found {
			t.Errorf("field %q not found", tt.name)
		}
	}
}

func TestDefaultProvisionerProducesEnvVars(t *testing.T) {
	ct := APIToken()

	if ct.DefaultProvisioner == nil {
		t.Fatal("DefaultProvisioner is nil")
	}

	in := sdk.ProvisionInput{
		ItemFields: map[sdk.FieldName]string{
			fieldToken:          "tok_abc",
			fieldGraph:          "my-graph",
			fieldAPIURL:         "https://custom.api/graph",
			fieldTimeoutSeconds: "15",
		},
	}

	out := sdk.ProvisionOutput{Environment: make(map[string]string)}
	ct.DefaultProvisioner.Provision(context.Background(), in, &out)

	env := out.Environment
	if env["ROAM_API_TOKEN"] != "tok_abc" {
		t.Errorf("ROAM_API_TOKEN = %q, want %q", env["ROAM_API_TOKEN"], "tok_abc")
	}
	if env["ROAM_API_GRAPH"] != "my-graph" {
		t.Errorf("ROAM_API_GRAPH = %q, want %q", env["ROAM_API_GRAPH"], "my-graph")
	}
	if env["ROAM_API_BASE_URL"] != "https://custom.api/graph" {
		t.Errorf("ROAM_API_BASE_URL = %q, want %q", env["ROAM_API_BASE_URL"], "https://custom.api/graph")
	}
	if env["ROAM_TIMEOUT_SECONDS"] != "15" {
		t.Errorf("ROAM_TIMEOUT_SECONDS = %q, want %q", env["ROAM_TIMEOUT_SECONDS"], "15")
	}
}

func TestDefaultProvisionerIsEnvVars(t *testing.T) {
	ct := APIToken()
	if ct.DefaultProvisioner == nil {
		t.Fatal("DefaultProvisioner is nil")
	}
	desc := ct.DefaultProvisioner.Description()
	if desc == "" {
		t.Error("DefaultProvisioner description is empty")
	}
}

func TestImporterImportsEnvVars(t *testing.T) {
	ct := APIToken()

	t.Setenv("ROAM_API_TOKEN", "tok_imported")
	t.Setenv("ROAM_API_GRAPH", "imported-graph")
	t.Setenv("ROAM_API_BASE_URL", "https://imported.api/graph")
	t.Setenv("ROAM_TIMEOUT_SECONDS", "42")

	out := sdk.ImportOutput{}
	ct.Importer(context.Background(), sdk.ImportInput{}, &out)

	candidates := out.AllCandidates()
	if len(candidates) == 0 {
		t.Fatal("expected at least 1 import candidate")
	}

	cand := candidates[0]
	if cand.Fields[fieldToken] != "tok_imported" {
		t.Errorf("Token = %q, want %q", cand.Fields[fieldToken], "tok_imported")
	}
	if cand.Fields[fieldGraph] != "imported-graph" {
		t.Errorf("Graph = %q, want %q", cand.Fields[fieldGraph], "imported-graph")
	}
	if cand.Fields[fieldAPIURL] != "https://imported.api/graph" {
		t.Errorf("API URL = %q, want %q", cand.Fields[fieldAPIURL], "https://imported.api/graph")
	}
	if cand.Fields[fieldTimeoutSeconds] != "42" {
		t.Errorf("Timeout Seconds = %q, want %q", cand.Fields[fieldTimeoutSeconds], "42")
	}
}
