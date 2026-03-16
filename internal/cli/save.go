package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"roam-cli/internal/format"
	"roam-cli/internal/roam"
)

func validateSaveTarget(title, parentUID, dailyPage string) error {
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
	if set == 0 {
		return fmt.Errorf("one of --title, --parent, or --to-daily-page is required")
	}
	if set > 1 {
		return fmt.Errorf("--title, --parent, and --to-daily-page are mutually exclusive")
	}
	return nil
}

func newSaveCmd() *cobra.Command {
	var title string
	var parentUID string
	var dailyPage string
	var file string
	var useStdin bool
	var pageUID string
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:     "save",
		Aliases: []string{"save-markdown"},
		Short:   "Save markdown as a Roam page or under a parent block",
		Example: "  cat note.md | roam-cli save --title \"New Page\"\n  cat note.md | roam-cli save --to-daily-page 2026-03-14\n  roam-cli save --parent <block-uid> --file ./note.md",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			if err := validateSaveTarget(title, parentUID, dailyPage); err != nil {
				return err
			}

			// Resolve --to-daily-page into --title
			if strings.TrimSpace(dailyPage) != "" {
				when, err := parseDateFlexible(dailyPage)
				if err != nil {
					return fmt.Errorf("invalid date for --to-daily-page: %w", err)
				}
				title = roam.DailyTitle(when)
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

			target := strings.TrimSpace(parentUID)
			actions := []map[string]any{}
			mode := "parent"
			if strings.TrimSpace(title) != "" {
				mode = "page"
				if pageUID == "" {
					pageUID = roam.NewUID()
				}
				actions = append(actions, roam.CreatePageAction(title, pageUID))
				target = pageUID
			}
			actions = append(actions, format.GFMToBatchActions(raw, target)...)

			client, err := mustClient()
			if err != nil {
				return err
			}
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

	cmd.Flags().StringVar(&title, "title", "", "Page title (required unless --parent or --to-daily-page is set)")
	cmd.Flags().StringVar(&parentUID, "parent", "", "Target parent block UID")
	cmd.Flags().StringVar(&dailyPage, "to-daily-page", "", "Save to daily page by date (e.g. 2026-03-14, defaults to today)")
	cmd.Flags().StringVar(&file, "file", "", "Markdown file path (default: stdin)")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read markdown from stdin")
	cmd.Flags().StringVar(&pageUID, "uid", "", "Optional page uid (only with --title)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output save result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output save result as plain text")
	return cmd
}
