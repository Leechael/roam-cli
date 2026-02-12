package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check roam-cli credential and API connectivity status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := mustClient()
			if err != nil {
				msg := "Roam credentials are not configured. Set ROAM_API_TOKEN and ROAM_API_GRAPH, then try again."
				if asJSON {
					return prettyPrint(map[string]any{"ok": false, "message": msg, "error": err.Error()})
				}
				return fmt.Errorf("%s\n%s", msg, err)
			}

			_, err = client.Q(`[:find ?title :where [?e :node/title ?title]]`, nil)
			if err != nil {
				msg := "Roam API check failed. Verify token, graph name, and network connectivity."
				if asJSON {
					return prettyPrint(map[string]any{"ok": false, "message": msg, "error": err.Error()})
				}
				return fmt.Errorf("%s\n%s", msg, err)
			}

			if asJSON {
				return prettyPrint(map[string]any{
					"ok":      true,
					"message": "roam-cli is configured and API is reachable",
				})
			}
			fmt.Println("OK: roam-cli is configured and Roam API is reachable.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output status as JSON")
	return cmd
}
