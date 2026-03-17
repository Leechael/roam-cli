package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	batchdsl "roam-cli/internal/batch"
)

func newBatchCmd() *cobra.Command {
	batch := &cobra.Command{
		Use:   "batch",
		Short: "Low-level batch actions API",
	}

	var file string
	var useStdin bool
	var asJSON bool
	var asPlain bool

	run := &cobra.Command{
		Use:     "run",
		Short:   "Run Roam batch-actions from JSON array",
		Long:    "Run Roam batch-actions from JSON array. Supports native actions (create-block, update-block, delete-block, move-block) and children/attach-to in create-block.",
		Example: "  roam-cli batch run --file ./examples/actions.bulk-update.json\n  roam-cli batch run --file ./examples/actions.create-with-children.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			raw, err := readAllFromFileOrStdin(file, useStdin || file == "")
			if err != nil {
				return err
			}
			raw = strings.TrimSpace(raw)
			if raw == "" {
				return nil
			}

			var actions []map[string]any
			if err := json.Unmarshal([]byte(raw), &actions); err != nil {
				return err
			}

			client, err := mustClient()
			if err != nil {
				return err
			}

			expanded, err := batchdsl.ExpandActions(actions, client)
			if err != nil {
				return err
			}

			resp, err := client.BatchActions(expanded)
			if err != nil {
				return err
			}
			if asJSON {
				return prettyPrint(map[string]any{"input_actions": len(actions), "expanded_actions": len(expanded), "response": resp})
			}
			fmt.Printf("ok (%d actions; expanded from %d)\n", len(expanded), len(actions))
			return nil
		},
	}

	run.Flags().StringVar(&file, "file", "", "Path to actions JSON file")
	run.Flags().BoolVar(&useStdin, "stdin", false, "Read actions JSON from stdin")
	run.Flags().BoolVar(&asJSON, "json", false, "Output batch result as JSON")
	run.Flags().BoolVar(&asPlain, "plain", false, "Output batch result as plain text")

	batch.AddCommand(run)
	return batch
}
