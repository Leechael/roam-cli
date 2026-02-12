package roam

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type SearchResult struct {
	UID       string `json:"uid"`
	Text      string `json:"text"`
	PageTitle string `json:"page_title"`
}

func escapeDatalogString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}

func toSlice(v any) []any {
	if v == nil {
		return nil
	}
	if s, ok := v.([]any); ok {
		return s
	}
	return nil
}

func toMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]any); ok {
		return m
	}
	return nil
}

func (c *Client) GetPageByTitle(title string) (map[string]any, error) {
	escaped := escapeDatalogString(title)
	query := fmt.Sprintf(`
[:find (pull ?e [*
                 {:block/children ...}
                 {:block/refs [*]}
                ])
 :where [?e :node/title "%s"]]
`, escaped)
	result, err := c.Q(query, nil)
	if err != nil {
		return nil, err
	}
	rows := toSlice(result)
	if len(rows) == 0 {
		return nil, nil
	}
	first := toSlice(rows[0])
	if len(first) == 0 {
		return nil, nil
	}
	return toMap(first[0]), nil
}

func (c *Client) GetBlockByUID(uid string) (map[string]any, error) {
	query := fmt.Sprintf(`
[:find (pull ?e [*
                 {:block/children ...}
                 {:block/refs [*]}
                ])
 :where [?e :block/uid "%s"]]
`, uid)
	result, err := c.Q(query, nil)
	if err != nil {
		return nil, err
	}
	rows := toSlice(result)
	if len(rows) == 0 {
		return nil, nil
	}
	first := toSlice(rows[0])
	if len(first) == 0 {
		return nil, nil
	}
	return toMap(first[0]), nil
}

func (c *Client) SearchBlocks(terms []string, limit int, caseSensitive bool, pageTitle string) ([]SearchResult, error) {
	if len(terms) == 0 {
		return []SearchResult{}, nil
	}
	conditions := make([]string, 0, len(terms))
	for _, term := range terms {
		escaped := escapeDatalogString(term)
		if caseSensitive {
			conditions = append(conditions, fmt.Sprintf(`[(clojure.string/includes? ?s "%s")]`, escaped))
		} else {
			conditions = append(conditions, fmt.Sprintf(`[(clojure.string/includes? (clojure.string/lower-case ?s) "%s")]`, strings.ToLower(escaped)))
		}
	}
	termConditions := strings.Join(conditions, "\n    ")

	var query string
	if pageTitle != "" {
		escapedTitle := escapeDatalogString(pageTitle)
		query = fmt.Sprintf(`
[:find ?uid ?s
 :where
    [?p :node/title "%s"]
    [?b :block/page ?p]
    [?b :block/string ?s]
    [?b :block/uid ?uid]
    %s]
`, escapedTitle, termConditions)
	} else {
		query = fmt.Sprintf(`
[:find ?uid ?s ?page-title
 :where
    [?b :block/string ?s]
    [?b :block/uid ?uid]
    [?b :block/page ?p]
    [?p :node/title ?page-title]
    %s]
`, termConditions)
	}

	result, err := c.Q(query, nil)
	if err != nil {
		return nil, err
	}

	rows := toSlice(result)
	out := make([]SearchResult, 0, len(rows))
	for _, r := range rows {
		cols := toSlice(r)
		if pageTitle != "" {
			if len(cols) < 2 {
				continue
			}
			out = append(out, SearchResult{
				UID:       fmt.Sprintf("%v", cols[0]),
				Text:      fmt.Sprintf("%v", cols[1]),
				PageTitle: pageTitle,
			})
		} else {
			if len(cols) < 3 {
				continue
			}
			out = append(out, SearchResult{
				UID:       fmt.Sprintf("%v", cols[0]),
				Text:      fmt.Sprintf("%v", cols[1]),
				PageTitle: fmt.Sprintf("%v", cols[2]),
			})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		pi, ti, pgi := relevance(out[i], terms, caseSensitive)
		pj, tj, pgj := relevance(out[j], terms, caseSensitive)
		if pi != pj {
			return pi < pj
		}
		if pgi != pgj {
			return pgi < pgj
		}
		return ti < tj
	})

	if limit <= 0 || limit >= len(out) {
		return out, nil
	}
	return out[:limit], nil
}

func relevance(item SearchResult, terms []string, caseSensitive bool) (priority int, text string, page string) {
	text = item.Text
	page = item.PageTitle
	for _, term := range terms {
		termCheck := term
		pageCheck := page
		textCheck := text
		if !caseSensitive {
			termCheck = strings.ToLower(term)
			pageCheck = strings.ToLower(page)
			textCheck = strings.ToLower(text)
		}
		if pageCheck == termCheck {
			return 0, text, page
		}
		if strings.Contains(textCheck, "[["+termCheck+"]]") || strings.Contains(textCheck, "#[["+termCheck+"]]") {
			return 1, text, page
		}
	}
	return 2, text, page
}

func (c *Client) GetJournalingByDate(day time.Time, topicNode string) ([]map[string]any, error) {
	dateUID := day.Format("01-02-2006")

	var query string
	if topicNode != "" {
		escapedTopic := escapeDatalogString(topicNode)
		query = fmt.Sprintf(`
[:find (pull ?e [*
                 {:block/children [*]}
                 {:block/refs [*]}
                ])
 :where
    [?id :block/string "%s"]
    [?id :block/parents ?pid]
    [?pid :block/uid "%s"]
    [?e :block/parents ?id]
    [?e :block/parents ?pid]
]
`, escapedTopic, dateUID)
	} else {
		query = fmt.Sprintf(`
[:find (pull ?e [*
                 {:block/children [*]}
                 {:block/refs [*]}
                ])
 :where
    [?pid :block/uid "%s"]
    [?e :block/parents ?id]
    [?e :block/parents ?pid]
]
`, dateUID)
	}

	result, err := c.Q(query, nil)
	if err != nil {
		return nil, err
	}
	rows := toSlice(result)
	if len(rows) == 0 {
		return []map[string]any{}, nil
	}
	var nodes []map[string]any
	for _, row := range rows {
		cols := toSlice(row)
		for _, col := range cols {
			if m := toMap(col); m != nil {
				nodes = append(nodes, m)
			}
		}
	}
	return nodes, nil
}
