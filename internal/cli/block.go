package cli

import (
	"encoding/json"
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
	block.AddCommand(newBlockFindCmd())
	block.AddCommand(newBlockCreateTreeCmd())

	return block
}

func newBlockCreateCmd() *cobra.Command {
	var parentUID, text, uid, order string
	var open bool
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a block",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
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
			if asPlain {
				fmt.Printf("%v\n", uidOut)
				return nil
			}
			return prettyPrint(map[string]any{"uid": uidOut, "response": resp})
		},
	}

	cmd.Flags().StringVar(&parentUID, "parent", "", "Parent uid")
	cmd.Flags().StringVar(&text, "text", "", "Block text")
	cmd.Flags().StringVar(&uid, "uid", "", "Optional uid")
	cmd.Flags().StringVar(&order, "order", "last", "Order: first|last|<int>")
	cmd.Flags().BoolVar(&open, "open", true, "Block open state")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output create result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output create result as plain text")

	return cmd
}

func newBlockUpdateCmd() *cobra.Command {
	var uid, text string
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a block text",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
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
			if asPlain {
				fmt.Printf("%s\n", uid)
				return nil
			}
			return prettyPrint(map[string]any{"uid": uid, "response": resp})
		},
	}

	cmd.Flags().StringVar(&uid, "uid", "", "Block uid")
	cmd.Flags().StringVar(&text, "text", "", "New block text")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output update result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output update result as plain text")

	return cmd
}

func newBlockDeleteCmd() *cobra.Command {
	var uid string
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a block",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
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
			if asPlain {
				fmt.Printf("%s\n", uid)
				return nil
			}
			return prettyPrint(map[string]any{"uid": uid, "response": resp})
		},
	}

	cmd.Flags().StringVar(&uid, "uid", "", "Block uid")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output delete result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output delete result as plain text")
	return cmd
}

func newBlockGetCmd() *cobra.Command {
	var uid string
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get block recursively by uid",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
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
			if asPlain {
				fmt.Printf("%v\n", result)
				return nil
			}
			return prettyPrint(result)
		},
	}

	cmd.Flags().StringVar(&uid, "uid", "", "Block uid")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output block data as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output block data as plain text")
	return cmd
}

func newBlockFindCmd() *cobra.Command {
	var text string
	var daily string
	var page string

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find a block UID by text match",
		RunE: func(cmd *cobra.Command, args []string) error {
			if text == "" {
				return errMissingFlag("text")
			}
			if daily != "" && page != "" {
				return fmt.Errorf("--daily and --page cannot be used together")
			}
			if daily == "" && page == "" {
				return fmt.Errorf("one of --daily or --page is required")
			}

			client, err := mustClient()
			if err != nil {
				return err
			}

			if daily != "" {
				when, err := parseDateFlexible(daily)
				if err != nil {
					return err
				}
				uid, err := client.FindBlockUID(text, "", &when)
				if err != nil {
					return err
				}
				fmt.Print(uid)
				return nil
			}

			uid, err := client.FindBlockUID(text, page, nil)
			if err != nil {
				return err
			}
			fmt.Print(uid)
			return nil
		},
	}

	cmd.Flags().StringVar(&text, "text", "", "Block text to match (required)")
	cmd.Flags().StringVar(&daily, "daily", "", "Daily note date (mutually exclusive with --page)")
	cmd.Flags().StringVar(&page, "page", "", "Page title (mutually exclusive with --daily)")
	return cmd
}

type treeNode struct {
	Text     string     `json:"text"`
	Children []treeNode `json:"children,omitempty"`
}

func newBlockCreateTreeCmd() *cobra.Command {
	var parentUID string
	var file string
	var useStdin bool
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "create-tree",
		Short: "Create a nested tree of blocks atomically",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			if parentUID == "" {
				return errMissingFlag("parent")
			}

			raw, err := readAllFromFileOrStdin(file, useStdin)
			if err != nil {
				return err
			}

			var nodes []treeNode
			raw = trimBOM(raw)
			// Try array first, then single object
			if err := json.Unmarshal([]byte(raw), &nodes); err != nil {
				var single treeNode
				if err2 := json.Unmarshal([]byte(raw), &single); err2 != nil {
					return fmt.Errorf("invalid JSON input: %w", err)
				}
				nodes = []treeNode{single}
			}

			var actions []map[string]any
			var walkTree func(parent string, children []treeNode)
			walkTree = func(parent string, children []treeNode) {
				for _, node := range children {
					uid := roam.NewUID()
					actions = append(actions, roam.CreateBlockAction(node.Text, parent, uid, "last", true))
					if len(node.Children) > 0 {
						walkTree(uid, node.Children)
					}
				}
			}
			walkTree(parentUID, nodes)

			if len(actions) == 0 {
				return fmt.Errorf("no blocks to create")
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
				fmt.Println("ok")
				return nil
			}
			return prettyPrint(resp)
		},
	}

	cmd.Flags().StringVar(&parentUID, "parent", "", "Parent block UID (required)")
	cmd.Flags().StringVar(&file, "file", "", "JSON input file")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read JSON from stdin")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output result as plain text")
	return cmd
}

func trimBOM(s string) string {
	if len(s) >= 3 && s[0] == 0xEF && s[1] == 0xBB && s[2] == 0xBF {
		return s[3:]
	}
	return s
}
