package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"roam-cli/internal/format"
	"roam-cli/internal/roam"
)

func validateSaveTarget(title, parentUID, pageUID string) error {
	t := strings.TrimSpace(title)
	p := strings.TrimSpace(parentUID)
	if t == "" && p == "" {
		return fmt.Errorf("one of --title or --parent is required")
	}
	if t != "" && p != "" {
		return fmt.Errorf("--title and --parent cannot be used together")
	}
	if p != "" && strings.TrimSpace(pageUID) != "" {
		return fmt.Errorf("--uid can only be used with --title")
	}
	return nil
}

func newSaveCmd() *cobra.Command {
	var title string
	var parentUID string
	var file string
	var useStdin bool
	var pageUID string
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:     "save",
		Aliases: []string{"save-markdown"},
		Short:   "Save markdown as a Roam page or under a parent block",
		Example: "  cat note.md | roam-cli save --title \"New Page\"\n  roam-cli save --title \"New Page\" --file ./note.md\n  roam-cli save --parent <block-uid> --file ./note.md",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			if err := validateSaveTarget(title, parentUID, pageUID); err != nil {
				return err
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

	cmd.Flags().StringVar(&title, "title", "", "Page title (required unless --parent is set)")
	cmd.Flags().StringVar(&parentUID, "parent", "", "Target parent block UID (required unless --title is set)")
	cmd.Flags().StringVar(&file, "file", "", "Markdown file path (default: stdin)")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read markdown from stdin")
	cmd.Flags().StringVar(&pageUID, "uid", "", "Optional page uid (only with --title)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output save result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output save result as plain text")
	return cmd
}
