package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newQCmd() *cobra.Command {
	var args []string
	var asJSON bool
	var asPlain bool
	var jqExpr string

	cmd := &cobra.Command{
		Use:   "q [query]",
		Short: "Execute raw datalog query",
		RunE: func(cmd *cobra.Command, commandArgs []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			if jqExpr != "" && !asJSON {
				return fmt.Errorf("--jq requires --json")
			}

			var queryArg string
			if len(commandArgs) > 0 {
				queryArg = commandArgs[0]
			}
			query, err := readQueryArgOrStdin(queryArg)
			if err != nil {
				return err
			}

			client, err := mustClient()
			if err != nil {
				return err
			}

			result, err := client.Q(query, args)
			if err != nil {
				return err
			}

			if jqExpr != "" {
				filtered, err := applyJQ(result, jqExpr)
				if err != nil {
					return err
				}
				return prettyPrint(filtered)
			}
			if asJSON {
				return prettyPrint(result)
			}
			fmt.Printf("%v\n", result)
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&args, "args", nil, "Query args")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output query results as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output query results as plain text")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter JSON output using a jq expression (requires --json)")
	return cmd
}
