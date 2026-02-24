package cli

import "testing"

func TestParseTreeNodes(t *testing.T) {
	raw := `{"string":"root","children":[{"text":"child"}]}`
	nodes, err := parseTreeNodes(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].blockText() != "root" {
		t.Fatalf("expected root text from string field")
	}
	if len(nodes[0].Children) != 1 || nodes[0].Children[0].blockText() != "child" {
		t.Fatalf("unexpected child parse")
	}
}

func TestParseTreeNodesArray(t *testing.T) {
	raw := `[{"text":"a"},{"string":"b"}]`
	nodes, err := parseTreeNodes(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
}
