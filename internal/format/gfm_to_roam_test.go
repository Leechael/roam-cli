package format

import (
	"strings"
	"testing"
)

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

func TestGFMToBatchActions_CodeFenceInsideListItem(t *testing.T) {
	md := "- Section title\n    - ```python\n@dataclass\nclass RaTLS:\n    x: int = 0\n\n    @classmethod\n    async def from_cert(cls): ...\n```\n"
	actions := GFMToBatchActions(md, "pageuid")
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d: %#v", len(actions), actions)
	}

	codeAction := actions[1]
	codeBlock := codeAction["block"].(map[string]any)
	codeText := codeBlock["string"].(string)
	if !strings.HasPrefix(codeText, "```python\n") {
		t.Fatalf("expected fenced code block, got: %q", codeText)
	}
	if !strings.Contains(codeText, "@dataclass") || !strings.Contains(codeText, "@classmethod") {
		t.Fatalf("expected code contents to stay intact, got: %q", codeText)
	}
	parentUID := codeAction["location"].(map[string]any)["parent-uid"].(string)
	firstUID := actions[0]["block"].(map[string]any)["uid"].(string)
	if parentUID != firstUID {
		t.Fatalf("expected code block parent %q, got %q", firstUID, parentUID)
	}
}

func TestGFMToBatchActions_CodeFenceKeepsFollowingListItems(t *testing.T) {
	md := "- **Section A**\n    - ```python\ncode here```\n    - item after code block\n- **Section B**\n    - should appear as next sibling\n"
	actions := GFMToBatchActions(md, "pageuid")
	got := []string{}
	for _, action := range actions {
		block, ok := action["block"].(map[string]any)
		if !ok {
			continue
		}
		if s, ok := block["string"].(string); ok {
			got = append(got, s)
		}
	}
	want := []string{"**Section A**", "```python\ncode here\n```", "item after code block", "**Section B**", "should appear as next sibling"}
	for _, needle := range want {
		found := false
		for _, s := range got {
			if strings.Contains(s, needle) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected to find %q in actions, got %#v", needle, got)
		}
	}
}
