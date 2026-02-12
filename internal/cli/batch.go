package cli

import (
	"encoding/json"
	"strings"

	"github.com/spf13/cobra"
)

func newBatchCmd() *cobra.Command {
	batch := &cobra.Command{
		Use:   "batch",
		Short: "Low-level batch actions API",
	}

	var file string
	var useStdin bool

	run := &cobra.Command{
		Use:   "run",
		Short: "Run Roam batch-actions from JSON array",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := readAllFromFileOrStdin(file, useStdin)
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
			resp, err := client.BatchActions(actions)
			if err != nil {
				return err
			}
			return prettyPrint(resp)
		},
	}

	run.Flags().StringVar(&file, "file", "", "Path to actions JSON file")
	run.Flags().BoolVar(&useStdin, "stdin", false, "Read actions JSON from stdin")

	batch.AddCommand(run)
	return batch
}
