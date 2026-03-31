package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Leechael/roam-cli/internal/client"
)

func newSearchCmd() *cobra.Command {
	var page string
	var ignoreCase bool
	var limit int
	var searchType string
	var dailyTopic string
	var dailyDepth int
	var estimate bool
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "search <terms...>",
		Short: "Search blocks or pages by terms",
		Long: `Search for content across your Roam graph.

  --type page    (default) Aggregate results by page, sorted by match count.
                 Each argument is an independent query; results are deduplicated.
  --type block   Return individual matching blocks, sorted by relevance.
                 All arguments are AND-matched as a single query.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if asJSON && asPlain {
				return fmt.Errorf("--json and --plain cannot be used together")
			}
			if searchType != "page" && searchType != "block" {
				return fmt.Errorf("--type must be \"page\" or \"block\"")
			}

			c, err := mustClient()
			if err != nil {
				return err
			}

			resolvedPage := maybeResolveDailyTitle(page)
			caseSensitive := !ignoreCase

			if searchType == "block" {
				return searchBlocks(c, args, limit, caseSensitive, resolvedPage, asJSON, asPlain)
			}
			return searchPages(cmd, c, args, limit, caseSensitive, resolvedPage, dailyTopic, dailyDepth, estimate, asJSON, asPlain)
		},
	}

	cmd.Flags().StringVar(&searchType, "type", "page", "Result type: page or block")
	cmd.Flags().StringVar(&page, "page", "", "Limit to page title or date")
	cmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "Case-insensitive search")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum results (0 = unlimited)")
	cmd.Flags().StringVar(&dailyTopic, "daily-topic", "", "Filter daily page sections by topic (--type page only)")
	cmd.Flags().IntVar(&dailyDepth, "daily-depth", 1, "Aggregation depth for daily pages (--type page only)")
	cmd.Flags().BoolVar(&estimate, "estimate", false, "Only show estimated counts (--type page only)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output as plain text")
	return cmd
}

func searchBlocks(c *client.Client, args []string, limit int, caseSensitive bool, page string, asJSON, asPlain bool) error {
	if limit == 0 {
		limit = 20
	}
	results, err := c.SearchBlocks(args, limit, caseSensitive, page)
	if err != nil {
		return err
	}
	if asJSON {
		return prettyPrint(results)
	}
	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	byPage := map[string][]struct{ uid, text string }{}
	order := []string{}
	for _, r := range results {
		if _, ok := byPage[r.PageTitle]; !ok {
			byPage[r.PageTitle] = nil
			order = append(order, r.PageTitle)
		}
		byPage[r.PageTitle] = append(byPage[r.PageTitle], struct{ uid, text string }{r.UID, r.Text})
	}

	total := 0
	for _, p := range order {
		fmt.Printf("[[%s]]\n", p)
		for _, b := range byPage[p] {
			t := strings.ReplaceAll(b.text, "\n", " ")
			if len(t) > 80 {
				t = t[:77] + "..."
			}
			fmt.Printf("  %s   %s\n", b.uid, t)
			total++
		}
		fmt.Println()
	}

	fmt.Printf("---\n%d blocks across %d pages", total, len(order))
	if total >= limit {
		fmt.Printf(" (limited to %d, use --limit to change)", limit)
	}
	fmt.Printf("\nUse --type page for results aggregated by page.\n")
	return nil
}

func searchPages(cmd *cobra.Command, c *client.Client, args []string, limit int, caseSensitive bool, page, dailyTopic string, dailyDepth int, estimate, asJSON, asPlain bool) error {
	if estimate {
		estimates, estFailed, err := c.EstimateSearch(args, caseSensitive, page)
		if err != nil {
			return err
		}
		if asJSON {
			out := map[string]any{"estimates": estimates}
			if len(estFailed) > 0 {
				out["failed_queries"] = estFailed
			}
			return prettyPrint(out)
		}
		if len(estFailed) > 0 {
			for _, f := range estFailed {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: skipped query: %s\n", f)
			}
			fmt.Fprintln(cmd.ErrOrStderr())
		}
		for _, e := range estimates {
			fmt.Printf("%q: %d blocks, %d pages\n", e.Query, e.BlockCount, e.PageCount)
		}
		return nil
	}

	results, failed, err := c.SearchPages(args, caseSensitive, page, dailyTopic, dailyDepth)
	if err != nil {
		return err
	}

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	if asJSON {
		out := map[string]any{
			"pages": results,
		}
		if len(failed) > 0 {
			out["failed_queries"] = failed
		}
		return prettyPrint(out)
	}

	if len(failed) > 0 {
		for _, f := range failed {
			fmt.Fprintf(cmd.ErrOrStderr(), "warning: skipped query: %s\n", f)
		}
		fmt.Fprintln(cmd.ErrOrStderr())
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	totalHits := 0
	for _, r := range results {
		totalHits += r.HitCount
	}

	for _, r := range results {
		if r.SectionUID != "" {
			fmt.Printf("((%s))\n", r.SectionUID)
			fmt.Printf("[[%s]] > %s\n", r.PageTitle, strings.ReplaceAll(r.SectionTitle, "\n", " "))
		} else {
			fmt.Printf("((%s))\n", r.PageUID)
			fmt.Printf("[[%s]]\n", r.PageTitle)
		}
		fmt.Printf("%d hits | matched: %s\n",
			r.HitCount,
			strings.Join(quoteAll(r.QueriesMatched), ", "))
		fmt.Println()
	}

	fmt.Printf("---\n%d pages, %d blocks matched across %d queries", len(results), totalHits, len(args))
	if limit > 0 && limit <= len(results) {
		fmt.Printf(" (limited to %d)", limit)
	}
	fmt.Printf("\nUse --type block for individual block matches.\n")

	return nil
}

func quoteAll(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = fmt.Sprintf("%q", s)
	}
	return out
}
