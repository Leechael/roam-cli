package batch

import (
	"fmt"
	"testing"
)

type mockFinder struct {
	results map[string]string // key: "text|parentUID" → uid
}

func (m *mockFinder) FindBlockUnderParent(text, parentUID string) (string, error) {
	key := fmt.Sprintf("%s|%s", text, parentUID)
	if uid, ok := m.results[key]; ok {
		return uid, nil
	}
	return "", nil
}

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

	expanded, err := ExpandActions(actions, nil)
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

func TestExpandActions_CreateBlockWithChildren(t *testing.T) {
	actions := []map[string]any{
		{
			"action":   "create-block",
			"location": map[string]any{"parent-uid": "p1", "order": "last"},
			"block": map[string]any{
				"string":   "parent",
				"children": []any{map[string]any{"text": "child"}},
			},
		},
	}

	expanded, err := ExpandActions(actions, nil)
	if err != nil {
		t.Fatalf("ExpandActions error: %v", err)
	}
	if len(expanded) != 2 {
		t.Fatalf("expected 2 expanded actions, got %d", len(expanded))
	}
	rootUID := expanded[0]["block"].(map[string]any)["uid"].(string)
	childLoc := expanded[1]["location"].(map[string]any)
	if childLoc["parent-uid"] != rootUID {
		t.Fatalf("expected child parent %s, got %v", rootUID, childLoc["parent-uid"])
	}
}

func TestExpandActions_CreateBlockWithAttachTo_Found(t *testing.T) {
	finder := &mockFinder{results: map[string]string{
		"[[📽 Journaling]]|page-uid": "existing-uid",
	}}
	actions := []map[string]any{
		{
			"action":   "create-block",
			"location": map[string]any{"parent-uid": "page-uid", "attach-to": "[[📽 Journaling]]", "order": "last"},
			"block":    map[string]any{"string": "new child"},
		},
	}

	expanded, err := ExpandActions(actions, finder)
	if err != nil {
		t.Fatalf("ExpandActions error: %v", err)
	}
	if len(expanded) != 1 {
		t.Fatalf("expected 1 action, got %d", len(expanded))
	}
	loc := expanded[0]["location"].(map[string]any)
	if loc["parent-uid"] != "existing-uid" {
		t.Fatalf("expected parent existing-uid, got %v", loc["parent-uid"])
	}
}

func TestExpandActions_CreateBlockWithAttachTo_NotFound(t *testing.T) {
	finder := &mockFinder{results: map[string]string{}}
	actions := []map[string]any{
		{
			"action":   "create-block",
			"location": map[string]any{"parent-uid": "page-uid", "attach-to": "[[📽 Journaling]]", "order": "last"},
			"block":    map[string]any{"string": "new child"},
		},
	}

	expanded, err := ExpandActions(actions, finder)
	if err != nil {
		t.Fatalf("ExpandActions error: %v", err)
	}
	// Should create attach-to block + the child
	if len(expanded) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(expanded))
	}
	// First action creates the attach-to block
	if expanded[0]["block"].(map[string]any)["string"] != "[[📽 Journaling]]" {
		t.Fatalf("expected attach-to block text, got %v", expanded[0]["block"])
	}
	attachUID := expanded[0]["block"].(map[string]any)["uid"].(string)
	// Second action creates the child under the attach-to block
	childLoc := expanded[1]["location"].(map[string]any)
	if childLoc["parent-uid"] != attachUID {
		t.Fatalf("expected child parent %s, got %v", attachUID, childLoc["parent-uid"])
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
	_, err := ExpandActions(actions, nil)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestExpandActions_PassThrough(t *testing.T) {
	actions := []map[string]any{{"action": "update-block", "block": map[string]any{"uid": "u1", "string": "x"}}}
	expanded, err := ExpandActions(actions, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expanded) != 1 || expanded[0]["action"] != "update-block" {
		t.Fatalf("expected passthrough update-block")
	}
}

func TestExpandActions_CreateBlockWithoutChildren_PassThrough(t *testing.T) {
	actions := []map[string]any{
		{
			"action":   "create-block",
			"location": map[string]any{"parent-uid": "p1", "order": "last"},
			"block":    map[string]any{"string": "simple block"},
		},
	}
	expanded, err := ExpandActions(actions, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(expanded) != 1 {
		t.Fatalf("expected 1 passthrough action, got %d", len(expanded))
	}
	if expanded[0]["action"] != "create-block" {
		t.Fatalf("expected create-block passthrough")
	}
}
