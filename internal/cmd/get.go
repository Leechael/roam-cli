package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Leechael/roam-cli/internal/client"
	"github.com/Leechael/roam-cli/internal/format"
	"github.com/Leechael/roam-cli/internal/parser"
)

func newGetCmd() *cobra.Command {
	var asJSON bool
	var asPlain bool
	var daily string
	var today bool

	cmd := &cobra.Command{
		Use:   "get [identifier]",
		Short: "Get page by title, block by uid, or daily page by date",
		Example: `  roam-cli get "Page Title"
  roam-cli get "((block-uid))"
  roam-cli get --today
  roam-cli get --daily yesterday
  roam-cli get --daily 2026-03-14`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if asJSON && asPlain {
				return fmt.Errorf("--json and --plain cannot be used together")
			}
			if today && daily != "" {
				return fmt.Errorf("--today and --daily cannot be used together")
			}
			if today {
				daily = "today"
			}

			hasDaily := daily != ""
			hasArg := len(args) > 0

			if hasDaily && hasArg {
				return fmt.Errorf("--today/--daily and positional argument are mutually exclusive")
			}
			if !hasDaily && !hasArg {
				return fmt.Errorf("provide a page title/block uid, or use --today/--daily")
			}

			c, err := mustClient()
			if err != nil {
				return err
			}

			// Resolve daily page to title
			var identifier string
			isPage := false
			if hasDaily {
				when, err := parseDateFlexible(daily)
				if err != nil {
					return err
				}
				identifier = client.DailyTitle(when)
			} else {
				identifier = args[0]
			}

			var result map[string]any
			if !hasDaily {
				if uid, ok := parser.ParseUID(identifier); ok {
					result, err = c.GetBlockByUID(uid)
					if err != nil {
						return err
					}
				}
			}
			if result == nil {
				result, err = c.GetPageByTitle(identifier)
				if err != nil {
					return err
				}
				isPage = true
			}
			if result == nil {
				return fmt.Errorf("not found: %s", identifier)
			}

			if asJSON {
				return prettyPrint(result)
			}

			if isPage {
				children := mapChildren(result)
				fmt.Println(format.FormatBlocksAsMarkdown(children))
				return nil
			}

			fmt.Println(format.FormatBlocksAsMarkdown([]map[string]any{result}))
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output raw JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output plain markdown")
	cmd.Flags().BoolVar(&today, "today", false, "Get today's daily page")
	cmd.Flags().StringVar(&daily, "daily", "", "Get daily page by date (YYYY-MM-DD, today, yesterday, tomorrow)")
	return cmd
}

func mapChildren(block map[string]any) []map[string]any {
	raw, ok := block[":block/children"]
	if !ok || raw == nil {
		return []map[string]any{}
	}
	arr, ok := raw.([]any)
	if !ok {
		return []map[string]any{}
	}
	out := make([]map[string]any, 0, len(arr))
	for _, v := range arr {
		if m, ok := v.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}
