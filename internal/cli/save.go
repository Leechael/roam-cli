package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"roam-cli/internal/format"
	"roam-cli/internal/roam"
)

func validateSaveTarget(title, parentUID, dailyPage string, today bool) error {
	set := 0
	if strings.TrimSpace(title) != "" {
		set++
	}
	if strings.TrimSpace(parentUID) != "" {
		set++
	}
	if strings.TrimSpace(dailyPage) != "" {
		set++
	}
	if today {
		set++
	}
	if set == 0 {
		return fmt.Errorf("one of --title, --parent, --to-daily-page, or --today is required")
	}
	if set > 1 {
		return fmt.Errorf("--title, --parent, --to-daily-page, and --today are mutually exclusive")
	}
	return nil
}

func newSaveCmd() *cobra.Command {
	var title string
	var parentUID string
	var dailyPage string
	var today bool
	var under string
	var file string
	var useStdin bool
	var pageUID string
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:     "save",
		Aliases: []string{"save-markdown"},
		Short:   "Save markdown as a Roam page or under a parent block",
		Example: `  cat note.md | roam-cli save --title "New Page"
  cat note.md | roam-cli save --to-daily-page 2026-03-14
  echo "- journal entry" | roam-cli save --today
  echo "- content" | roam-cli save --today --under '[[📽 Journaling]]'
  roam-cli save --parent <block-uid> --file ./note.md`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			if err := validateSaveTarget(title, parentUID, dailyPage, today); err != nil {
				return err
			}

			// --today is shorthand for --to-daily-page with today's date
			if today {
				dailyPage = time.Now().Format("2006-01-02")
			}

			// Resolve --to-daily-page into title
			if strings.TrimSpace(dailyPage) != "" {
				when, err := parseDateFlexible(dailyPage)
				if err != nil {
					return fmt.Errorf("invalid date for --to-daily-page: %w", err)
				}
				title = roam.DailyTitle(when)
			}

			// --under requires a page target, not --parent
			if strings.TrimSpace(under) != "" && strings.TrimSpace(title) == "" {
				return fmt.Errorf("--under requires --title, --to-daily-page, or --today")
			}

			if strings.TrimSpace(pageUID) != "" && strings.TrimSpace(title) == "" {
				return fmt.Errorf("--uid can only be used with --title")
			}

			raw, err := readAllFromFileOrStdin(file, useStdin || file == "")
			if err != nil {
				return err
			}
			if strings.TrimSpace(raw) == "" {
				return fmt.Errorf("no markdown content provided")
			}

			// Need client early for page/block lookups
			client, err := mustClient()
			if err != nil {
				return err
			}

			target := strings.TrimSpace(parentUID)
			actions := []map[string]any{}
			mode := "parent"

			if strings.TrimSpace(title) != "" {
				mode = "page"

				// Upsert: check if page already exists
				existingUID, err := client.GetPageUIDByTitle(title)
				if err != nil {
					return fmt.Errorf("failed to check existing page: %w", err)
				}
				if existingUID != "" {
					// Page exists — append to it
					pageUID = existingUID
				} else {
					// Page does not exist — create it
					if pageUID == "" {
						pageUID = roam.NewUID()
					}
					actions = append(actions, roam.CreatePageAction(title, pageUID))
				}
				target = pageUID
			}

			// --under: find-or-create a direct child block under the page
			if strings.TrimSpace(under) != "" {
				foundUID, err := client.FindBlockUnderParent(under, target)
				if err != nil {
					return fmt.Errorf("--under lookup failed: %w", err)
				}
				if foundUID != "" {
					target = foundUID
				} else {
					newUID := roam.NewUID()
					actions = append(actions, roam.CreateBlockAction(under, target, newUID, "last", true))
					target = newUID
				}
			}

			actions = append(actions, format.GFMToBatchActions(raw, target)...)

			resp, err := client.BatchActions(actions)
			if err != nil {
				return err
			}

			payload := map[string]any{
				"mode":       mode,
				"title":      title,
				"page_uid":   pageUID,
				"parent_uid": parentUID,
				"actions":    len(actions),
				"response":   resp,
			}
			if asJSON {
				return prettyPrint(payload)
			}
			if mode == "page" {
				fmt.Printf("saved page %q (%s) with %d actions\n", title, pageUID, len(actions))
				return nil
			}
			fmt.Printf("saved markdown under parent %q with %d actions\n", parentUID, len(actions))
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Page title (required unless --parent, --to-daily-page, or --today is set)")
	cmd.Flags().StringVar(&parentUID, "parent", "", "Target parent block UID")
	cmd.Flags().StringVar(&dailyPage, "to-daily-page", "", "Save to daily page by date (e.g. 2026-03-14)")
	cmd.Flags().BoolVar(&today, "today", false, "Save to today's daily page")
	cmd.Flags().StringVar(&under, "under", "", "Find-or-create a child block with this text under the target page, then append content under it")
	cmd.Flags().StringVar(&file, "file", "", "Markdown file path (default: stdin)")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read markdown from stdin")
	cmd.Flags().StringVar(&pageUID, "uid", "", "Optional page uid (only with --title)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output save result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output save result as plain text")
	return cmd
}
