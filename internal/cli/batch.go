package cli

import (
	"encoding/json"
	"fmt"
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
	var asJSON bool
	var asPlain bool

	run := &cobra.Command{
		Use:   "run",
		Short: "Run Roam batch-actions from JSON array",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
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
			if asPlain {
				fmt.Printf("ok (%d actions)\n", len(actions))
				return nil
			}
			return prettyPrint(resp)
		},
	}

	run.Flags().StringVar(&file, "file", "", "Path to actions JSON file")
	run.Flags().BoolVar(&useStdin, "stdin", false, "Read actions JSON from stdin")
	run.Flags().BoolVar(&asJSON, "json", false, "Output batch result as JSON")
	run.Flags().BoolVar(&asPlain, "plain", false, "Output batch result as plain text")

	batch.AddCommand(run)
	return batch
}
