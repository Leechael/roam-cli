package batch

import "testing"

func TestExpandActions_CreateWithChildren(t *testing.T) {
	actions := []map[string]any{
		{
			"action":   "create-with-children",
			"location": map[string]any{"parent-uid": "parent-1", "order": "first"},
			"block": map[string]any{
				"string": "root",
				"children": []any{
					map[string]any{"text": "child-a"},
					map[string]any{"string": "child-b"},
				},
			},
		},
	}

	expanded, err := ExpandActions(actions)
	if err != nil {
		t.Fatalf("ExpandActions error: %v", err)
	}
	if len(expanded) != 3 {
		t.Fatalf("expected 3 expanded actions, got %d", len(expanded))
	}
	if expanded[0]["action"] != "create-block" {
		t.Fatalf("expected first action create-block, got %#v", expanded[0]["action"])
	}
	loc0 := expanded[0]["location"].(map[string]any)
	if loc0["parent-uid"] != "parent-1" {
		t.Fatalf("expected root parent parent-1, got %v", loc0["parent-uid"])
	}
	if loc0["order"] != "first" {
		t.Fatalf("expected root order first, got %v", loc0["order"])
	}
	rootUID := expanded[0]["block"].(map[string]any)["uid"].(string)
	if rootUID == "" {
		t.Fatalf("expected generated root uid")
	}
	loc1 := expanded[1]["location"].(map[string]any)
	if loc1["parent-uid"] != rootUID {
		t.Fatalf("expected child parent %s, got %v", rootUID, loc1["parent-uid"])
	}
}

func TestExpandActions_InvalidCreateWithChildren(t *testing.T) {
	actions := []map[string]any{
		{
			"action":   "create-with-children",
			"location": map[string]any{"parent-uid": "p"},
			"block": map[string]any{
				"string": "",
			},
		},
	}
	_, err := ExpandActions(actions)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestExpandActions_PassThrough(t *testing.T) {
	actions := []map[string]any{{"action": "update-block", "block": map[string]any{"uid": "u1", "string": "x"}}}
	expanded, err := ExpandActions(actions)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expanded) != 1 || expanded[0]["action"] != "update-block" {
		t.Fatalf("expected passthrough update-block")
	}
}
