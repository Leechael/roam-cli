package model

import "fmt"

// APIError represents an error response from the Roam API.
type APIError struct {
	Status int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("roam api error: status=%d body=%s", e.Status, e.Body)
}

// SearchResult holds a single block match from a search query.
type SearchResult struct {
	UID       string `json:"uid"`
	Text      string `json:"text"`
	PageTitle string `json:"page_title"`
	PageUID   string `json:"page_uid,omitempty"`
}

// PageSearchResult holds search results aggregated by page.
type PageSearchResult struct {
	PageTitle      string   `json:"page_title"`
	PageUID        string   `json:"page_uid"`
	SectionTitle   string   `json:"section_title,omitempty"`
	SectionUID     string   `json:"section_uid,omitempty"`
	HitCount       int      `json:"hit_count"`
	QueriesMatched []string `json:"queries_matched"`
}

// EstimateResult holds block and page counts for a single query.
type EstimateResult struct {
	Query      string `json:"query"`
	BlockCount int    `json:"block_count"`
	PageCount  int    `json:"page_count"`
}
