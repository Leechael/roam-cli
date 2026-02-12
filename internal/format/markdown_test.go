package format

import "testing"

func TestFormatBlocksAsMarkdown(t *testing.T) {
	blocks := []map[string]any{
		{
			":block/string":  "Title",
			":block/heading": 2,
			":block/children": []any{
				map[string]any{":block/string": "child line", ":block/order": 0},
			},
		},
	}
	out := FormatBlocksAsMarkdown(blocks)
	if out == "" {
		t.Fatal("expected non-empty markdown")
	}
	if out[:2] != "##" {
		t.Fatalf("expected heading markdown, got: %s", out)
	}
}
