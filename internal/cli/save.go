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

	cmd := &cobra.Command{
		Use:     "save",
		Aliases: []string{"save-markdown"},
		Short:   "Save markdown as a Roam page",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			return prettyPrint(map[string]any{
				"title":    title,
				"page_uid": pageUID,
				"actions":  len(actions),
				"response": resp,
			})
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "Page title")
	cmd.Flags().StringVar(&file, "file", "", "Markdown file path (default: stdin)")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read markdown from stdin")
	cmd.Flags().StringVar(&pageUID, "uid", "", "Optional page uid")
	return cmd
}
