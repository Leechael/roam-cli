package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"roam-cli/internal/format"
	"roam-cli/internal/roam"
)

func newSaveCmd() *cobra.Command {
	var title string
	var file string
	var useStdin bool
	var pageUID string
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:     "save",
		Aliases: []string{"save-markdown"},
		Short:   "Save markdown as a Roam page",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			if strings.TrimSpace(title) == "" {
				return errMissingFlag("title")
			}
			raw, err := readAllFromFileOrStdin(file, useStdin || file == "")
			if err != nil {
				return err
			}
			if strings.TrimSpace(raw) == "" {
				return fmt.Errorf("no markdown content provided")
			}
			if pageUID == "" {
				pageUID = roam.NewUID()
			}

			actions := []map[string]any{roam.CreatePageAction(title, pageUID)}
			actions = append(actions, format.GFMToBatchActions(raw, pageUID)...)

			client, err := mustClient()
			if err != nil {
				return err
			}
			resp, err := client.BatchActions(actions)
			if err != nil {
				return err
			}

			payload := map[string]any{
				"title":    title,
				"page_uid": pageUID,
				"actions":  len(actions),
				"response": resp,
			}
			if asJSON {
				return prettyPrint(payload)
			}
			fmt.Printf("saved page %q (%s) with %d actions\n", title, pageUID, len(actions))
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Page title")
	cmd.Flags().StringVar(&file, "file", "", "Markdown file path (default: stdin)")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read markdown from stdin")
	cmd.Flags().StringVar(&pageUID, "uid", "", "Optional page uid")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output save result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output save result as plain text")
	return cmd
}
