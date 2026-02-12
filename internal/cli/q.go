package cli

import "github.com/spf13/cobra"

func newQCmd() *cobra.Command {
	var args []string

	cmd := &cobra.Command{
		Use:   "q [query]",
		Short: "Execute raw datalog query",
		RunE: func(cmd *cobra.Command, commandArgs []string) error {
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
			return prettyPrint(result)
		},
	}

	cmd.Flags().StringArrayVar(&args, "args", nil, "Query args")
	return cmd
}
