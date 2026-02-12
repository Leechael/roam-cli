package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	var page string
	var ignoreCase bool
	var limit int

	cmd := &cobra.Command{
		Use:   "search <terms...>",
		Short: "Search blocks containing all terms",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				return err
			}

			results, err := client.SearchBlocks(args, limit, !ignoreCase, page)
			if err != nil {
				return err
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

			for _, p := range order {
				fmt.Printf("[[%s]]\n", p)
				for _, b := range byPage[p] {
					t := strings.ReplaceAll(b.text, "\n", " ")
					if len(t) > 80 {
						t = t[:77] + "..."
					}
					fmt.Printf("  %s   %s\n", b.uid, t)
				}
				fmt.Println()
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&page, "page", "", "Limit to page title")
	cmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "Case-insensitive search")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum results")
	return cmd
}
