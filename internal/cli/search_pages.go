package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newSearchPagesCmd() *cobra.Command {
	var page string
	var dailyTopic string
	var dailyDepth int
	var ignoreCase bool
	var limit int
	var estimate bool
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "search-pages <query>...",
		Short: "Multi-query search aggregated by page",
		Long: `Run multiple independent searches and return results grouped by page.

Each positional argument is an independent query. Results are deduplicated,
aggregated by page, and sorted by number of queries matched then hit count.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}

			if estimate {
				estimates, estFailed, err := client.EstimateSearch(args, !ignoreCase, maybeResolveDailyTitle(page))
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

			results, failed, err := client.SearchPages(args, !ignoreCase, maybeResolveDailyTitle(page), dailyTopic, dailyDepth)
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

			fmt.Printf("---\n%d pages, %d blocks matched across %d queries\n",
				len(results), totalHits, len(args))

			return nil
		},
	}

	cmd.Flags().StringVar(&page, "page", "", "Limit to page title or date")
	cmd.Flags().StringVar(&dailyTopic, "daily-topic", "", "Filter daily page sections by topic node text (at depth-1 level)")
	cmd.Flags().IntVar(&dailyDepth, "daily-depth", 1, "Aggregation depth for daily pages (1=top-level children, 2=grandchildren, etc.)")
	cmd.Flags().BoolVar(&estimate, "estimate", false, "Only show estimated block/page counts per query")
	cmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "Case-insensitive search")
	cmd.Flags().IntVar(&limit, "limit", 0, "Maximum pages to return (0 = unlimited)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func quoteAll(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = fmt.Sprintf("%q", s)
	}
	return out
}
