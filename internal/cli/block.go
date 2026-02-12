package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"roam-cli/internal/roam"
)

func newBlockCmd() *cobra.Command {
	block := &cobra.Command{
		Use:   "block",
		Short: "Low-level block APIs",
	}

	block.AddCommand(newBlockCreateCmd())
	block.AddCommand(newBlockUpdateCmd())
	block.AddCommand(newBlockDeleteCmd())
	block.AddCommand(newBlockGetCmd())

	return block
}

func newBlockCreateCmd() *cobra.Command {
	var parentUID, text, uid, order string
	var open bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a block",
		RunE: func(cmd *cobra.Command, args []string) error {
			if parentUID == "" {
				return errMissingFlag("parent")
			}
			if text == "" {
				return errMissingFlag("text")
			}

			client, err := mustClient()
			if err != nil {
				return err
			}

			action := roam.CreateBlockAction(text, parentUID, uid, order, open)
			resp, err := client.Write(action)
			if err != nil {
				return err
			}

			uidOut := action["block"].(map[string]any)["uid"]
			return prettyPrint(map[string]any{"uid": uidOut, "response": resp})
		},
	}

	cmd.Flags().StringVar(&parentUID, "parent", "", "Parent uid")
	cmd.Flags().StringVar(&text, "text", "", "Block text")
	cmd.Flags().StringVar(&uid, "uid", "", "Optional uid")
	cmd.Flags().StringVar(&order, "order", "last", "Order: first|last|<int>")
	cmd.Flags().BoolVar(&open, "open", true, "Block open state")

	return cmd
}

func newBlockUpdateCmd() *cobra.Command {
	var uid, text string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a block text",
		RunE: func(cmd *cobra.Command, args []string) error {
			if uid == "" {
				return errMissingFlag("uid")
			}
			if text == "" {
				return errMissingFlag("text")
			}
			client, err := mustClient()
			if err != nil {
				return err
			}
			resp, err := client.Write(roam.UpdateBlockAction(uid, text))
			if err != nil {
				return err
			}
			return prettyPrint(map[string]any{"uid": uid, "response": resp})
		},
	}

	cmd.Flags().StringVar(&uid, "uid", "", "Block uid")
	cmd.Flags().StringVar(&text, "text", "", "New block text")

	return cmd
}

func newBlockDeleteCmd() *cobra.Command {
	var uid string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a block",
		RunE: func(cmd *cobra.Command, args []string) error {
			if uid == "" {
				return errMissingFlag("uid")
			}
			client, err := mustClient()
			if err != nil {
				return err
			}
			resp, err := client.Write(roam.DeleteBlockAction(uid))
			if err != nil {
				return err
			}
			return prettyPrint(map[string]any{"uid": uid, "response": resp})
		},
	}

	cmd.Flags().StringVar(&uid, "uid", "", "Block uid")
	return cmd
}

func newBlockGetCmd() *cobra.Command {
	var uid string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get block recursively by uid",
		RunE: func(cmd *cobra.Command, args []string) error {
			if uid == "" {
				return errMissingFlag("uid")
			}
			client, err := mustClient()
			if err != nil {
				return err
			}
			query := fmt.Sprintf(`
[:find (pull ?e [*
                 {:block/children ...}
                 {:block/refs [*]}
                ])
 :where [?e :block/uid "%s"]]
`, uid)
			result, err := client.Q(query, nil)
			if err != nil {
				return err
			}
			return prettyPrint(result)
		},
	}

	cmd.Flags().StringVar(&uid, "uid", "", "Block uid")
	return cmd
}
