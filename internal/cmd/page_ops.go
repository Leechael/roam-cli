package cmd

import (
	"fmt"
	"strings"

	"github.com/Leechael/roam-cli/internal/client"
)

func resolvePageTarget(args []string, daily string, today bool) (string, error) {
	if today && strings.TrimSpace(daily) != "" {
		return "", fmt.Errorf("--today and --daily cannot be used together")
	}
	if today {
		daily = "today"
	}
	if strings.TrimSpace(daily) != "" {
		if len(args) > 0 {
			return "", fmt.Errorf("--today/--daily and positional argument are mutually exclusive")
		}
		when, err := parseDateFlexible(daily)
		if err != nil {
			return "", err
		}
		return client.DailyTitle(when), nil
	}
	if len(args) != 1 {
		return "", fmt.Errorf("provide exactly one page title, or use --today/--daily")
	}
	return strings.TrimSpace(args[0]), nil
}

func clearPageContent(c *client.Client, title string) (string, int, map[string]any, error) {
	page, err := c.GetPageByTitle(title)
	if err != nil {
		return "", 0, nil, err
	}
	if page == nil {
		return "", 0, nil, fmt.Errorf("not found: %s", title)
	}
	pageUID := fmt.Sprintf("%v", page[":block/uid"])
	children := mapChildren(page)
	if len(children) == 0 {
		return pageUID, 0, map[string]any{"ok": true}, nil
	}

	actions := make([]map[string]any, 0, len(children))
	for _, child := range children {
		uid := fmt.Sprintf("%v", child[":block/uid"])
		if strings.TrimSpace(uid) == "" {
			continue
		}
		actions = append(actions, client.DeleteBlockAction(uid))
	}
	if len(actions) == 0 {
		return pageUID, 0, map[string]any{"ok": true}, nil
	}

	resp, err := c.BatchActions(actions)
	if err != nil {
		return "", 0, nil, err
	}
	return pageUID, len(actions), resp, nil
}

func deletePageByTitle(c *client.Client, title string) (string, map[string]any, error) {
	pageUID, err := c.GetPageUIDByTitle(title)
	if err != nil {
		return "", nil, err
	}
	if strings.TrimSpace(pageUID) == "" {
		return "", nil, fmt.Errorf("not found: %s", title)
	}
	resp, err := c.Write(client.DeletePageAction(pageUID))
	if err != nil {
		return "", nil, err
	}
	return pageUID, resp, nil
}
