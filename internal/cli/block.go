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

	return block
}

func newBlockCreateCmd() *cobra.Command {
	var parentUID, text, uid, order string
	var attachTo string
	var file string
	var useStdin bool
	var open bool
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a block or nested tree of blocks",
		Long: `Create blocks under a parent UID. Supports three modes:

1. Single block:   --text "content"
2. Nested tree:    --file tree.json  (or pipe JSON via stdin)
3. Attach-to:      --attach-to "[[Section]]" finds or creates the section block
                   under --parent, then creates content under it.

JSON input accepts a single object or array; each node: text/string, children, uid, order, open.`,
		Example: `  roam-cli block create --parent <uid> --text "Hello"
  echo '{"text":"Root","children":[{"text":"Child"}]}' | roam-cli block create --parent <uid>
  roam-cli block create --parent <uid> --attach-to "[[📽 Journaling]]" --file items.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			if parentUID == "" {
				return errMissingFlag("parent")
			}

			client, err := mustClient()
			if err != nil {
				return err
			}

			// Resolve attach-to: find or create the intermediate parent
			effectiveParent := parentUID
			if attachTo != "" {
				foundUID, err := client.FindBlockUnderParent(attachTo, parentUID)
				if err != nil {
					return fmt.Errorf("attach-to lookup failed: %w", err)
				}
				if foundUID != "" {
					effectiveParent = foundUID
				} else {
					// Create the attach-to block
					newUID := roam.NewUID()
					action := roam.CreateBlockAction(attachTo, parentUID, newUID, order, true)
					if _, err := client.Write(action); err != nil {
						return fmt.Errorf("failed to create attach-to block: %w", err)
					}
					effectiveParent = newUID
				}
			}

			// Determine mode: JSON tree input vs single block
			hasFile := file != ""
			hasStdin := useStdin
			hasText := text != ""

			if hasFile || hasStdin {
				// Tree mode: JSON input
				raw, err := readAllFromFileOrStdin(file, hasStdin || file != "")
				if err != nil {
					return err
				}
				nodes, err := parseTreeNodes(raw)
				if err != nil {
					return err
				}

				var actions []map[string]any
				if err := walkTree(&actions, effectiveParent, nodes, "root"); err != nil {
					return err
				}
				if len(actions) == 0 {
					return fmt.Errorf("no blocks to create")
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
			}

			if !hasText {
				// Try reading stdin as JSON if no --text and no --file
				raw, err := readAllFromFileOrStdin("", true)
				if err == nil && strings.TrimSpace(raw) != "" {
					nodes, err := parseTreeNodes(raw)
					if err != nil {
						return errMissingFlag("text")
					}
					var actions []map[string]any
					if err := walkTree(&actions, effectiveParent, nodes, "root"); err != nil {
						return err
					}
					if len(actions) == 0 {
						return fmt.Errorf("no blocks to create")
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
				}
				return errMissingFlag("text")
			}

			// Single block mode
			action := roam.CreateBlockAction(text, effectiveParent, uid, order, open)
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

	cmd.Flags().StringVar(&parentUID, "parent", "", "Parent block UID (required)")
	cmd.Flags().StringVar(&text, "text", "", "Block text (single block mode)")
	cmd.Flags().StringVar(&uid, "uid", "", "Optional block uid")
	cmd.Flags().StringVar(&order, "order", "last", "Order: first|last|<int>")
	cmd.Flags().BoolVar(&open, "open", true, "Block open state")
	cmd.Flags().StringVar(&attachTo, "attach-to", "", "Find or create a block with this text under parent, use as actual parent")
	cmd.Flags().StringVar(&file, "file", "", "JSON input file for tree creation")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "Read JSON tree from stdin")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output result as plain text")

	return cmd
}

// walkTree recursively generates create-block actions from tree nodes.
func walkTree(actions *[]map[string]any, parent string, children []treeNode, path string) error {
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
		*actions = append(*actions, roam.CreateBlockAction(text, parent, uid, order, open))
		if len(node.Children) > 0 {
			if err := walkTree(actions, uid, node.Children, nodePath+".children"); err != nil {
				return err
			}
		}
	}
	return nil
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
