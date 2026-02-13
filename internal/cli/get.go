package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"roam-cli/internal/format"
	"roam-cli/internal/parser"
)

func newGetCmd() *cobra.Command {
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "get <identifier>",
		Short: "Get page by title or block by uid",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if asJSON && asPlain {
				return fmt.Errorf("--json and --plain cannot be used together")
			}

			identifier := args[0]
			client, err := mustClient()
			if err != nil {
				return err
			}

			var result map[string]any
			isPage := false
			if uid, ok := parser.ParseUID(identifier); ok {
				result, err = client.GetBlockByUID(uid)
				if err != nil {
					return err
				}
			}
			if result == nil {
				result, err = client.GetPageByTitle(identifier)
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
