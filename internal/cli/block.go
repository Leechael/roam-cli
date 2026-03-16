package cli

import (
	"encoding/json"
	"fmt"
	"strings"

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
	block.AddCommand(newBlockMoveCmd())
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

func newBlockMoveCmd() *cobra.Command {
	var uid string
	var parentUID string
	var order string
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move a block to another parent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			if strings.TrimSpace(uid) == "" {
				return errMissingFlag("uid")
			}
			if strings.TrimSpace(parentUID) == "" {
				return errMissingFlag("parent")
			}
			client, err := mustClient()
			if err != nil {
				return err
			}
			action := roam.MoveBlockAction(uid, parentUID, order)
			resp, err := client.Write(action)
			if err != nil {
				return err
			}
			if asPlain {
				fmt.Printf("%s\n", uid)
				return nil
			}
			return prettyPrint(map[string]any{"uid": uid, "parent_uid": parentUID, "response": resp})
		},
	}

	cmd.Flags().StringVar(&uid, "uid", "", "Block uid")
	cmd.Flags().StringVar(&parentUID, "parent", "", "Target parent uid")
	cmd.Flags().StringVar(&order, "order", "last", "Order: first|last|<int>")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output move result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output move result as plain text")
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

			uid, err := client.FindBlockUID(text, maybeResolveDailyTitle(page), nil)
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
	String   string     `json:"string"`
	UID      string     `json:"uid,omitempty"`
	Order    string     `json:"order,omitempty"`
	Open     *bool      `json:"open,omitempty"`
	Children []treeNode `json:"children,omitempty"`
}

func (n treeNode) blockText() string {
	if strings.TrimSpace(n.Text) != "" {
		return n.Text
	}
	return n.String
}

func newBlockCreateTreeCmd() *cobra.Command {
	var parentUID string
	var file string
	var useStdin bool
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:     "create-tree",
		Short:   "Create a nested tree of blocks atomically",
		Long:    "Create a nested tree under a parent block UID. Input JSON accepts either a single node object or an array of nodes. Each node supports text (or string), optional children, uid, order, and open.",
		Example: "  echo '{\"text\":\"Root\",\"children\":[{\"string\":\"Child\"}]}' | roam-cli block create-tree --parent <uid>\n  roam-cli block create-tree --parent <uid> --file ./tree.json",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			if parentUID == "" {
				return errMissingFlag("parent")
			}

			raw, err := readAllFromFileOrStdin(file, useStdin || file == "")
			if err != nil {
				return err
			}

			nodes, err := parseTreeNodes(raw)
			if err != nil {
				return err
			}

			var actions []map[string]any
			var walkTree func(parent string, children []treeNode, path string) error
			walkTree = func(parent string, children []treeNode, path string) error {
				for i, node := range children {
					nodePath := fmt.Sprintf("%s[%d]", path, i)
					text := strings.TrimSpace(node.blockText())
					if text == "" {
						return fmt.Errorf("%s text is required (use \"text\" or \"string\")", nodePath)
					}
					uid := strings.TrimSpace(node.UID)
					if uid == "" {
						uid = roam.NewUID()
					}
					order := strings.TrimSpace(node.Order)
					if order == "" {
						order = "last"
					}
					open := true
					if node.Open != nil {
						open = *node.Open
					}
					actions = append(actions, roam.CreateBlockAction(text, parent, uid, order, open))
					if len(node.Children) > 0 {
						if err := walkTree(uid, node.Children, nodePath+".children"); err != nil {
							return err
						}
					}
				}
				return nil
			}
			if err := walkTree(parentUID, nodes, "root"); err != nil {
				return err
			}

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

func parseTreeNodes(raw string) ([]treeNode, error) {
	raw = trimBOM(raw)
	var nodes []treeNode
	if err := json.Unmarshal([]byte(raw), &nodes); err == nil {
		return nodes, nil
	}
	var single treeNode
	if err := json.Unmarshal([]byte(raw), &single); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}
	return []treeNode{single}, nil
}

func trimBOM(s string) string {
	if len(s) >= 3 && s[0] == 0xEF && s[1] == 0xBB && s[2] == 0xBF {
		return s[3:]
	}
	return s
}
