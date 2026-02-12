package format

import "testing"

func TestGFMToBatchActions(t *testing.T) {
	md := "## Heading\n\n- item1\n- item2\n"
	actions := GFMToBatchActions(md, "pageuid")
	if len(actions) < 3 {
		t.Fatalf("expected at least 3 actions, got %d", len(actions))
	}

	first := actions[0]
	if first["action"] != "create-block" {
		t.Fatalf("unexpected first action: %#v", first)
	}
	block := first["block"].(map[string]any)
	if block["heading"] != 2 {
		t.Fatalf("expected heading 2, got %#v", block["heading"])
	}
}
