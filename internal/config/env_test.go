package config

import "testing"

func TestNewConfigFromArgs(t *testing.T) {
	cfg, err := New("token", "graph", "https://api.example.com/api/graph", 5)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Token != "token" || cfg.Graph != "graph" {
		t.Fatalf("unexpected token/graph: %#v", cfg)
	}
	if cfg.BaseURL != "https://api.example.com/api/graph" {
		t.Fatalf("unexpected base url: %s", cfg.BaseURL)
	}
	if cfg.TimeoutSeconds != 5 {
		t.Fatalf("unexpected timeout: %d", cfg.TimeoutSeconds)
	}
}

func TestNewConfigMissingRequired(t *testing.T) {
	_, err := New("", "", "", 0)
	if err == nil {
		t.Fatal("expected error for missing token/graph")
	}
}
