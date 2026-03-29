package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	var asJSON bool
	var asPlain bool
	var jqExpr string

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check roam-cli credential and API connectivity status",
		RunE: func(cmd *cobra.Command, args []string) error {
			if asJSON && asPlain {
				return fmt.Errorf("--json and --plain cannot be used together")
			}
			if jqExpr != "" && !asJSON {
				return fmt.Errorf("--jq requires --json")
			}

			render := func(payload map[string]any) error {
				if jqExpr != "" {
					filtered, err := applyJQ(payload, jqExpr)
					if err != nil {
						return err
					}
					return prettyPrint(filtered)
				}
				if asJSON {
					return prettyPrint(payload)
				}
				return nil
			}

			c, err := mustClient()
			if err != nil {
				msg := "Roam credentials are not configured. Set ROAM_API_TOKEN and ROAM_API_GRAPH, then try again."
				payload := map[string]any{"ok": false, "message": msg, "error": err.Error()}
				if asJSON {
					if printErr := render(payload); printErr != nil {
						return printErr
					}
				}
				return fmt.Errorf("%s\n%s", msg, err)
			}

			_, err = c.Q(`[:find ?title :where [?e :node/title ?title]]`, nil)
			if err != nil {
				msg := "Roam API check failed. Verify token, graph name, and network connectivity."
				payload := map[string]any{"ok": false, "message": msg, "error": err.Error()}
				if asJSON {
					if printErr := render(payload); printErr != nil {
						return printErr
					}
				}
				return fmt.Errorf("%s\n%s", msg, err)
			}

			payload := map[string]any{
				"ok":      true,
				"message": "roam-cli is configured and API is reachable",
			}
			if asJSON {
				return render(payload)
			}
			fmt.Println("OK: roam-cli is configured and Roam API is reachable.")
			return nil
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Output status as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output status as plain text")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output using a jq expression (requires --json)")
	return cmd
}
