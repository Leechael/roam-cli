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
	PageUID   string `json:"page_uid,omitempty"`
}

type PageSearchResult struct {
	PageTitle      string   `json:"page_title"`
	PageUID        string   `json:"page_uid"`
	SectionTitle   string   `json:"section_title,omitempty"`
	SectionUID     string   `json:"section_uid,omitempty"`
	HitCount       int      `json:"hit_count"`
	QueriesMatched []string `json:"queries_matched"`
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
[:find ?uid ?s ?page-uid
 :where
    [?p :node/title "%s"]
    [?p :block/uid ?page-uid]
    [?b :block/page ?p]
    [?b :block/string ?s]
    [?b :block/uid ?uid]
    %s]
`, escapedTitle, termConditions)
	} else {
		query = fmt.Sprintf(`
[:find ?uid ?s ?page-title ?page-uid
 :where
    [?b :block/string ?s]
    [?b :block/uid ?uid]
    [?b :block/page ?p]
    [?p :node/title ?page-title]
    [?p :block/uid ?page-uid]
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
			if len(cols) < 3 {
				continue
			}
			out = append(out, SearchResult{
				UID:       fmt.Sprintf("%v", cols[0]),
				Text:      fmt.Sprintf("%v", cols[1]),
				PageTitle: pageTitle,
				PageUID:   fmt.Sprintf("%v", cols[2]),
			})
		} else {
			if len(cols) < 4 {
				continue
			}
			out = append(out, SearchResult{
				UID:       fmt.Sprintf("%v", cols[0]),
				Text:      fmt.Sprintf("%v", cols[1]),
				PageTitle: fmt.Sprintf("%v", cols[2]),
				PageUID:   fmt.Sprintf("%v", cols[3]),
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

func (c *Client) FindBlockUID(text string, pageTitle string, dailyDate *time.Time) (string, error) {
	escaped := escapeDatalogString(text)
	var query string
	if dailyDate != nil {
		dateUID := dailyDate.Format("01-02-2006")
		query = fmt.Sprintf(`
[:find ?uid
 :where
   [?p :block/uid "%s"]
   [?b :block/page ?p]
   [?b :block/string "%s"]
   [?b :block/uid ?uid]]
`, dateUID, escaped)
	} else {
		escapedTitle := escapeDatalogString(pageTitle)
		query = fmt.Sprintf(`
[:find ?uid
 :where
   [?p :node/title "%s"]
   [?b :block/page ?p]
   [?b :block/string "%s"]
   [?b :block/uid ?uid]]
`, escapedTitle, escaped)
	}

	result, err := c.Q(query, nil)
	if err != nil {
		return "", err
	}
	rows := toSlice(result)
	if len(rows) == 0 {
		return "", fmt.Errorf("block not found: %s", text)
	}
	first := toSlice(rows[0])
	if len(first) == 0 {
		return "", fmt.Errorf("block not found: %s", text)
	}
	return fmt.Sprintf("%v", first[0]), nil
}

// FindBlockUnderParent finds a direct child block with matching text under the given parent UID.
// Returns the block UID if found, or empty string if not found.
func (c *Client) FindBlockUnderParent(text string, parentUID string) (string, error) {
	escaped := escapeDatalogString(text)
	escapedParent := escapeDatalogString(parentUID)
	query := fmt.Sprintf(`
[:find ?uid
 :where
   [?parent :block/uid "%s"]
   [?b :block/parents ?parent]
   [?b :block/string "%s"]
   [?b :block/uid ?uid]]
`, escapedParent, escaped)
	result, err := c.Q(query, nil)
	if err != nil {
		return "", err
	}
	rows := toSlice(result)
	if len(rows) == 0 {
		return "", nil
	}
	first := toSlice(rows[0])
	if len(first) == 0 {
		return "", nil
	}
	return fmt.Sprintf("%v", first[0]), nil
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

// isDailyPageTitle checks if a title matches Roam's daily page format:
// "January 1st, 2026", "February 25th, 2026", "March 8th, 2026", etc.
func isDailyPageTitle(title string) bool {
	months := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
	for _, m := range months {
		if strings.HasPrefix(title, m+" ") {
			// Check ends with ", YYYY"
			if len(title) >= len(m)+8 && title[len(title)-6] == ',' && title[len(title)-5] == ' ' {
				return true
			}
		}
	}
	return false
}

// buildDailyWhere generates Datalog where-clauses that walk from a page down
// to the target depth. The final level is bound to ?sec. If topic is non-empty,
// a substring filter is applied at level 1.
func buildDailyWhere(pageUID string, depth int, topic string) string {
	var clauses []string
	clauses = append(clauses, fmt.Sprintf(`[?p :block/uid "%s"]`, escapeDatalogString(pageUID)))

	prev := "?p"
	for i := 1; i <= depth; i++ {
		cur := "?sec"
		if i < depth {
			cur = fmt.Sprintf("?l%d", i)
		}
		clauses = append(clauses, fmt.Sprintf("[%s :block/children %s]", prev, cur))

		if topic != "" && i == 1 {
			textVar := cur + "-text"
			clauses = append(clauses, fmt.Sprintf("[%s :block/string %s]", cur, textVar))
			clauses = append(clauses, fmt.Sprintf(
				`[(clojure.string/includes? (clojure.string/lower-case %s) "%s")]`,
				textVar, strings.ToLower(escapeDatalogString(topic))))
		}

		prev = cur
	}
	return strings.Join(clauses, "\n    ")
}

// getDailySections returns (sectionUID -> sectionText) for nodes at the given
// depth under the daily page. If topic is non-empty, filters at level 1.
func (c *Client) getDailySections(pageUID string, depth int, topic string) (map[string]string, error) {
	where := buildDailyWhere(pageUID, depth, topic)
	query := fmt.Sprintf("[:find ?sec-uid ?sec-text\n :where\n    %s\n    [?sec :block/uid ?sec-uid]\n    [?sec :block/string ?sec-text]]", where)

	result, err := c.Q(query, nil)
	if err != nil {
		return nil, err
	}
	rows := toSlice(result)
	out := make(map[string]string, len(rows))
	for _, r := range rows {
		cols := toSlice(r)
		if len(cols) >= 2 {
			out[fmt.Sprintf("%v", cols[0])] = fmt.Sprintf("%v", cols[1])
		}
	}
	return out, nil
}

// getDailyAncestorMap returns (blockUID -> sectionUID) mapping each descendant
// to its ancestor section at the given depth.
func (c *Client) getDailyAncestorMap(pageUID string, depth int, topic string) (map[string]string, error) {
	where := buildDailyWhere(pageUID, depth, topic)
	query := fmt.Sprintf("[:find ?block-uid ?sec-uid\n :where\n    %s\n    [?sec :block/uid ?sec-uid]\n    [?b :block/parents ?sec]\n    [?b :block/uid ?block-uid]]", where)

	result, err := c.Q(query, nil)
	if err != nil {
		return nil, err
	}
	rows := toSlice(result)
	out := make(map[string]string, len(rows))
	for _, r := range rows {
		cols := toSlice(r)
		if len(cols) >= 2 {
			out[fmt.Sprintf("%v", cols[0])] = fmt.Sprintf("%v", cols[1])
		}
	}
	return out, nil
}

// SearchPages runs multiple independent queries, deduplicates blocks,
// and returns results aggregated by page. For daily pages, results are
// drilled down to the level specified by dailyDepth. If dailyTopic is
// non-empty, filters daily page sections at level 1.
func (c *Client) SearchPages(queries []string, caseSensitive bool, pageTitle string, dailyTopic string, dailyDepth int) ([]PageSearchResult, []string, error) {
	if len(queries) == 0 {
		return []PageSearchResult{}, nil, nil
	}

	// Phase 1: multi-query recall, collect all matching blocks
	type blockMatch struct {
		uid       string
		pageTitle string
		pageUID   string
		queries   map[string]bool
	}
	allBlocks := map[string]*blockMatch{}
	var failed []string

	for i, q := range queries {
		if i > 0 {
			time.Sleep(500 * time.Millisecond)
		}
		terms := strings.Fields(q)
		if len(terms) == 0 {
			continue
		}
		results, err := c.SearchBlocks(terms, 0, caseSensitive, pageTitle)
		if err != nil {
			failed = append(failed, fmt.Sprintf("%s (%v)", q, err))
			continue
		}
		for _, r := range results {
			bm, ok := allBlocks[r.UID]
			if !ok {
				bm = &blockMatch{
					uid:       r.UID,
					pageTitle: r.PageTitle,
					pageUID:   r.PageUID,
					queries:   map[string]bool{},
				}
				allBlocks[r.UID] = bm
			}
			bm.queries[q] = true
		}
	}

	// Phase 2: aggregate — named pages at page level, daily pages at section level
	type aggKey struct {
		pageTitle    string
		pageUID      string
		sectionTitle string
		sectionUID   string
	}
	type aggVal struct {
		hits    int
		queries map[string]bool
	}
	agg := map[aggKey]*aggVal{}

	addToAgg := func(key aggKey, bm *blockMatch) {
		v, ok := agg[key]
		if !ok {
			v = &aggVal{queries: map[string]bool{}}
			agg[key] = v
		}
		v.hits++
		for q := range bm.queries {
			v.queries[q] = true
		}
	}

	// Separate daily page blocks from named page blocks
	dailyPageBlocks := map[string][]*blockMatch{} // pageUID -> blocks
	for _, bm := range allBlocks {
		if isDailyPageTitle(bm.pageTitle) {
			dailyPageBlocks[bm.pageUID] = append(dailyPageBlocks[bm.pageUID], bm)
		} else {
			addToAgg(aggKey{pageTitle: bm.pageTitle, pageUID: bm.pageUID}, bm)
		}
	}

	// Phase 2b: resolve daily page sections at the configured depth
	if dailyDepth < 1 {
		dailyDepth = 1
	}
	for pageUID, blocks := range dailyPageBlocks {
		time.Sleep(500 * time.Millisecond)

		sections, secErr := c.getDailySections(pageUID, dailyDepth, dailyTopic)
		time.Sleep(500 * time.Millisecond)
		ancestors, ancErr := c.getDailyAncestorMap(pageUID, dailyDepth, dailyTopic)

		pgTitle := blocks[0].pageTitle

		if secErr != nil || ancErr != nil {
			// fallback: aggregate at page level
			for _, bm := range blocks {
				addToAgg(aggKey{pageTitle: pgTitle, pageUID: pageUID}, bm)
			}
			continue
		}

		for _, bm := range blocks {
			var secUID, secText string
			if text, ok := sections[bm.uid]; ok {
				// block IS a section node at the target depth
				secUID = bm.uid
				secText = text
			} else if topUID, ok := ancestors[bm.uid]; ok {
				secUID = topUID
				secText = sections[topUID]
			}

			if secUID != "" {
				addToAgg(aggKey{
					pageTitle:    pgTitle,
					pageUID:      pageUID,
					sectionTitle: secText,
					sectionUID:   secUID,
				}, bm)
			} else {
				// block is above target depth or filtered out by topic — skip
				continue
			}
		}
	}

	// Phase 3: build output
	out := make([]PageSearchResult, 0, len(agg))
	for key, val := range agg {
		matched := make([]string, 0, len(val.queries))
		for q := range val.queries {
			matched = append(matched, q)
		}
		sort.Strings(matched)
		out = append(out, PageSearchResult{
			PageTitle:      key.pageTitle,
			PageUID:        key.pageUID,
			SectionTitle:   key.sectionTitle,
			SectionUID:     key.sectionUID,
			HitCount:       val.hits,
			QueriesMatched: matched,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if len(out[i].QueriesMatched) != len(out[j].QueriesMatched) {
			return len(out[i].QueriesMatched) > len(out[j].QueriesMatched)
		}
		if out[i].HitCount != out[j].HitCount {
			return out[i].HitCount > out[j].HitCount
		}
		return out[i].PageTitle < out[j].PageTitle
	})

	return out, failed, nil
}
