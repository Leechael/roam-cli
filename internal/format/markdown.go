package format

import (
	"fmt"
	"sort"
	"strings"
)

func getString(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func getInt(m map[string]any, key string) int {
	v, ok := m[key]
	if !ok || v == nil {
		return 0
	}
	switch t := v.(type) {
	case int:
		return t
	case int32:
		return int(t)
	case int64:
		return int(t)
	case float64:
		return int(t)
	default:
		return 0
	}
}

func childrenOf(block map[string]any) []map[string]any {
	raw, ok := block[":block/children"]
	if !ok || raw == nil {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, v := range arr {
		if m, ok := v.(map[string]any); ok {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return getInt(out[i], ":block/order") < getInt(out[j], ":block/order")
	})
	return out
}

func flattenBlocks(block map[string]any, out *[]map[string]any) {
	*out = append(*out, block)
	for _, c := range childrenOf(block) {
		flattenBlocks(c, out)
	}
}

func extractRefs(text string) []string {
	refs := []string{}
	for {
		start := strings.Index(text, "((")
		if start < 0 {
			break
		}
		rest := text[start+2:]
		end := strings.Index(rest, "))")
		if end < 0 {
			break
		}
		uid := rest[:end]
		if uid != "" {
			refs = append(refs, uid)
		}
		text = rest[end+2:]
	}
	return refs
}

func resolveRefs(text string, all []map[string]any) string {
	refs := extractRefs(text)
	for _, ref := range refs {
		for _, b := range all {
			if getString(b, ":block/uid") == ref {
				text = strings.ReplaceAll(text, "(("+ref+"))", getString(b, ":block/string"))
				break
			}
		}
	}
	return text
}

func isTableBlock(block map[string]any) bool {
	return strings.TrimSpace(getString(block, ":block/string")) == "{{[[table]]}}"
}

func isCodeBlock(block map[string]any) bool {
	return strings.HasPrefix(strings.TrimSpace(getString(block, ":block/string")), "```")
}

func formatBlockText(block map[string]any, all []map[string]any) string {
	text := resolveRefs(getString(block, ":block/string"), all)
	heading := getInt(block, ":block/heading")
	if heading > 0 {
		if heading > 6 {
			heading = 6
		}
		return strings.Repeat("#", heading) + " " + text
	}
	return text
}

func collectChildrenFlat(block map[string]any, minLevel int, currentLevel int, out *[]map[string]any) {
	direct := childrenOf(block)
	childLevel := currentLevel + 1
	for _, child := range direct {
		if childLevel >= minLevel {
			*out = append(*out, child)
		}
		collectChildrenFlat(child, minLevel, childLevel, out)
	}
}

func collectRowCells(row map[string]any) []string {
	cells := []string{getString(row, ":block/string")}
	current := row
	for {
		children := childrenOf(current)
		if len(children) == 0 {
			break
		}
		current = children[0]
		cells = append(cells, getString(current, ":block/string"))
	}
	return cells
}

func formatTable(block map[string]any, all []map[string]any) string {
	rows := childrenOf(block)
	if len(rows) == 0 {
		return ""
	}
	table := make([][]string, 0, len(rows))
	maxCols := 0
	for _, row := range rows {
		cells := collectRowCells(row)
		for i := range cells {
			cells[i] = resolveRefs(cells[i], all)
		}
		if len(cells) > maxCols {
			maxCols = len(cells)
		}
		table = append(table, cells)
	}
	if len(table) == 0 || maxCols == 0 {
		return ""
	}
	for i := range table {
		for len(table[i]) < maxCols {
			table[i] = append(table[i], "")
		}
	}
	lines := []string{}
	lines = append(lines, "| "+strings.Join(table[0], " | ")+" |")
	sep := make([]string, maxCols)
	for i := range sep {
		sep[i] = "---"
	}
	lines = append(lines, "| "+strings.Join(sep, " | ")+" |")
	for i := 1; i < len(table); i++ {
		lines = append(lines, "| "+strings.Join(table[i], " | ")+" |")
	}
	return strings.Join(lines, "\n")
}

func FormatBlocksAsMarkdown(blocks []map[string]any) string {
	if len(blocks) == 0 {
		return ""
	}
	all := []map[string]any{}
	for _, b := range blocks {
		flattenBlocks(b, &all)
	}

	lines := []string{}
	for _, block := range blocks {
		if isTableBlock(block) {
			if len(lines) > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, formatTable(block, all))
			lines = append(lines, "")
			continue
		}

		text := formatBlockText(block, all)
		heading := getInt(block, ":block/heading")
		if heading > 0 {
			if len(lines) > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, text)
		} else {
			lines = append(lines, text)
		}
		lines = append(lines, "")

		level1 := childrenOf(block)
		for _, child := range level1 {
			if isTableBlock(child) {
				lines = append(lines, formatTable(child, all))
				lines = append(lines, "")
				continue
			}
			childText := formatBlockText(child, all)
			childHeading := getInt(child, ":block/heading")
			if childHeading > 0 {
				if len(lines) > 0 {
					lines = append(lines, "")
				}
				lines = append(lines, childText)
			} else {
				lines = append(lines, childText)
			}
			lines = append(lines, "")

			deep := []map[string]any{}
			collectChildrenFlat(child, 2, 1, &deep)
			if len(deep) > 0 {
				listItems := []string{}
				for _, d := range deep {
					dt := resolveRefs(getString(d, ":block/string"), all)
					if isCodeBlock(d) {
						if len(listItems) > 0 {
							lines = append(lines, listItems...)
							listItems = nil
							lines = append(lines, "")
						}
						lines = append(lines, dt)
						lines = append(lines, "")
					} else {
						listItems = append(listItems, "- "+dt)
					}
				}
				if len(listItems) > 0 {
					lines = append(lines, listItems...)
					lines = append(lines, "")
				}
			}
		}
	}

	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func FormatJournal(nodes []map[string]any, topicMode bool) string {
	if len(nodes) == 0 {
		return ""
	}
	roots := []map[string]any{}
	for _, n := range nodes {
		parentsLen := len(childrenAnyToSlice(n[":block/parents"]))
		if topicMode {
			if parentsLen == 2 {
				roots = append(roots, n)
			}
		} else {
			if parentsLen == 1 {
				roots = append(roots, n)
			}
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		return getInt(roots[i], ":block/order") < getInt(roots[j], ":block/order")
	})

	parts := make([]string, 0, len(roots))
	for _, r := range roots {
		parts = append(parts, formatBlockIndented(r, nodes, 0))
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func childrenAnyToSlice(v any) []any {
	if v == nil {
		return nil
	}
	if s, ok := v.([]any); ok {
		return s
	}
	return nil
}

func formatBlockIndented(b map[string]any, nodes []map[string]any, indent int) string {
	text := getString(b, ":block/string")
	text = resolveRefs(text, nodes)
	line := text
	if indent > 0 {
		line = strings.Repeat("=", indent) + "> " + text
	}

	childrenRaw := childrenAnyToSlice(b[":block/children"])
	if len(childrenRaw) == 0 {
		return line
	}

	children := []map[string]any{}
	for _, ch := range childrenRaw {
		cm, ok := ch.(map[string]any)
		if !ok {
			continue
		}
		dbID := cm[":db/id"]
		var real map[string]any
		for _, n := range nodes {
			if fmt.Sprintf("%v", n[":db/id"]) == fmt.Sprintf("%v", dbID) {
				real = n
				break
			}
		}
		if real != nil {
			children = append(children, real)
		}
	}
	sort.Slice(children, func(i, j int) bool {
		return getInt(children[i], ":block/order") < getInt(children[j], ":block/order")
	})

	lines := []string{line}
	for _, child := range children {
		lines = append(lines, formatBlockIndented(child, nodes, indent+2))
	}
	return strings.Join(lines, "\n")
}
