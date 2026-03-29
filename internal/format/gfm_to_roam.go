package format

import (
	"regexp"
	"strings"

	"github.com/Leechael/roam-cli/internal/client"
)

type parentFrame struct {
	level int
	uid   string
}

type listFrame struct {
	indent int
	uid    string
}

var (
	headingRe = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	listRe    = regexp.MustCompile(`^(\s*)([-*+]|\d+\.)\s+(.+)$`)
)

func GFMToBatchActions(raw, pageUID string) []map[string]any {
	lines := strings.Split(raw, "\n")
	actions := []map[string]any{}
	parents := []parentFrame{{level: 0, uid: pageUID}}
	listStack := []listFrame{}

	currentParent := func() string {
		return parents[len(parents)-1].uid
	}

	resetList := func() {
		listStack = nil
	}

	flushParagraph := func(buf *[]string) {
		if len(*buf) == 0 {
			return
		}
		text := strings.TrimSpace(strings.Join(*buf, "\n"))
		if text != "" {
			actions = append(actions, client.CreateBlockAction(text, currentParent(), "", "last", true))
		}
		*buf = nil
	}

	paragraph := []string{}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			flushParagraph(&paragraph)
			resetList()
			lang := strings.TrimPrefix(trimmed, "```")
			codeLines := []string{}
			for i = i + 1; i < len(lines); i++ {
				if strings.TrimSpace(lines[i]) == "```" {
					break
				}
				codeLines = append(codeLines, lines[i])
			}
			codeText := "```" + lang + "\n" + strings.Join(codeLines, "\n") + "\n```"
			actions = append(actions, client.CreateBlockAction(codeText, currentParent(), "", "last", true))
			continue
		}

		if isTableStart(lines, i) {
			flushParagraph(&paragraph)
			resetList()
			tableLines := []string{}
			for ; i < len(lines); i++ {
				if strings.TrimSpace(lines[i]) == "" {
					i--
					break
				}
				if !strings.Contains(lines[i], "|") {
					i--
					break
				}
				tableLines = append(tableLines, lines[i])
			}
			actions = append(actions, tableToActions(tableLines, currentParent())...)
			continue
		}

		if trimmed == "" {
			flushParagraph(&paragraph)
			resetList()
			continue
		}

		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			flushParagraph(&paragraph)
			resetList()
			continue
		}

		if m := headingRe.FindStringSubmatch(trimmed); len(m) == 3 {
			flushParagraph(&paragraph)
			resetList()
			level := len(m[1])
			text := strings.TrimSpace(m[2])
			if level == 1 {
				continue
			}
			if level > 3 {
				level = 3
			}
			for len(parents) > 0 && parents[len(parents)-1].level >= level {
				parents = parents[:len(parents)-1]
			}
			if len(parents) == 0 {
				parents = []parentFrame{{level: 0, uid: pageUID}}
			}
			action := client.CreateBlockAction(text, currentParent(), "", "last", true)
			action["block"].(map[string]any)["heading"] = level
			actions = append(actions, action)
			newUID := action["block"].(map[string]any)["uid"].(string)
			parents = append(parents, parentFrame{level: level, uid: newUID})
			continue
		}

		if m := listRe.FindStringSubmatch(line); len(m) == 4 {
			flushParagraph(&paragraph)
			indent := len(m[1])
			marker := m[2]
			text := strings.TrimSpace(m[3])
			if strings.HasSuffix(marker, ".") {
				text = marker + " " + text
			}

			for len(listStack) > 0 && indent <= listStack[len(listStack)-1].indent {
				listStack = listStack[:len(listStack)-1]
			}
			parentUID := currentParent()
			if len(listStack) > 0 {
				parentUID = listStack[len(listStack)-1].uid
			}
			action := client.CreateBlockAction(text, parentUID, "", "last", true)
			actions = append(actions, action)
			newUID := action["block"].(map[string]any)["uid"].(string)
			listStack = append(listStack, listFrame{indent: indent, uid: newUID})
			continue
		}

		if strings.HasPrefix(trimmed, "> ") {
			flushParagraph(&paragraph)
			resetList()
			actions = append(actions, client.CreateBlockAction(trimmed, currentParent(), "", "last", true))
			continue
		}

		paragraph = append(paragraph, line)
	}

	flushParagraph(&paragraph)
	return actions
}

func isTableStart(lines []string, i int) bool {
	if i+1 >= len(lines) {
		return false
	}
	line := strings.TrimSpace(lines[i])
	next := strings.TrimSpace(lines[i+1])
	if !strings.Contains(line, "|") || !strings.Contains(next, "|") {
		return false
	}
	return strings.Contains(next, "---")
}

func splitTableLine(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

func tableToActions(lines []string, parentUID string) []map[string]any {
	if len(lines) < 2 {
		return nil
	}
	actions := []map[string]any{}
	table := client.CreateBlockAction("{{[[table]]}}", parentUID, "", "last", false)
	actions = append(actions, table)
	tableUID := table["block"].(map[string]any)["uid"].(string)

	rows := []string{lines[0]}
	for i := 2; i < len(lines); i++ {
		rows = append(rows, lines[i])
	}

	for _, rowLine := range rows {
		cells := splitTableLine(rowLine)
		if len(cells) == 0 {
			continue
		}
		parent := tableUID
		for _, cell := range cells {
			cellAction := client.CreateBlockAction(cell, parent, "", "last", true)
			actions = append(actions, cellAction)
			parent = cellAction["block"].(map[string]any)["uid"].(string)
		}
	}
	return actions
}
