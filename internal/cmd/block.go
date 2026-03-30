package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Leechael/roam-cli/internal/client"
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
	var daily string
	var today bool
	var page string
	var file string
	var useStdin bool
	var open bool
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a block or nested tree of blocks",
		Long: `Create blocks under a parent. The parent can be specified by:

  --parent <uid>       Direct block/page UID
  --today              Today's daily page
  --daily <date>       Daily page by date (YYYY-MM-DD, today, yesterday, tomorrow)
  --page <title>       Page by title

Content modes:

  --text "content"     Single block
  --file tree.json     Nested tree from JSON file (or pipe via stdin)
  --attach-to "[[X]]"  Find or create section block, then create content under it

JSON input accepts a single object or array; each node: text/string, children, uid, order, open.`,
		Example: `  roam-cli block create --parent <uid> --text "Hello"
  roam-cli block create --today --attach-to "[[📽 Journaling]]" --text "item"
  roam-cli block create --daily yesterday --attach-to "[[📽 Journaling]]" --file items.json
  roam-cli block create --page "Project Notes" --text "todo"
  echo '{"text":"Root","children":[{"text":"Child"}]}' | roam-cli block create --parent <uid>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			if today && daily != "" {
				return fmt.Errorf("--today and --daily cannot be used together")
			}
			if today {
				daily = "today"
			}

			// Count how many parent-target flags are set
			targets := 0
			if parentUID != "" {
				targets++
			}
			if daily != "" {
				targets++
			}
			if page != "" {
				targets++
			}
			if targets == 0 {
				return fmt.Errorf("one of --parent, --today, --daily, or --page is required")
			}
			if targets > 1 {
				return fmt.Errorf("--parent, --today/--daily, and --page are mutually exclusive")
			}

			c, err := mustClient()
			if err != nil {
				return err
			}

			// Resolve --daily or --page into parentUID
			if daily != "" {
				when, err := parseDateFlexible(daily)
				if err != nil {
					return err
				}
				title := client.DailyTitle(when)
				pageUID, err := c.GetPageUIDByTitle(title)
				if err != nil {
					return fmt.Errorf("failed to look up daily page %q: %w", title, err)
				}
				if pageUID == "" {
					return fmt.Errorf("daily page %q does not exist", title)
				}
				parentUID = pageUID
			}
			if page != "" {
				pageUID, err := c.GetPageUIDByTitle(page)
				if err != nil {
					return fmt.Errorf("failed to look up page %q: %w", page, err)
				}
				if pageUID == "" {
					return fmt.Errorf("page %q does not exist", page)
				}
				parentUID = pageUID
			}

			// Resolve attach-to: find or create the intermediate parent
			effectiveParent := parentUID
			if attachTo != "" {
				foundUID, err := c.FindBlockUnderParent(attachTo, parentUID)
				if err != nil {
					return fmt.Errorf("attach-to lookup failed: %w", err)
				}
				if foundUID != "" {
					effectiveParent = foundUID
				} else {
					// Create the attach-to block
					newUID := client.NewUID()
					action := client.CreateBlockAction(attachTo, parentUID, newUID, order, true)
					if _, err := c.Write(action); err != nil {
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

				resp, err := c.BatchActions(actions)
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
					resp, err := c.BatchActions(actions)
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
			action := client.CreateBlockAction(text, effectiveParent, uid, order, open)
			resp, err := c.Write(action)
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

	cmd.Flags().StringVar(&parentUID, "parent", "", "Parent block/page UID")
	cmd.Flags().BoolVar(&today, "today", false, "Use today's daily page as parent")
	cmd.Flags().StringVar(&daily, "daily", "", "Use daily page as parent (YYYY-MM-DD, today, yesterday, tomorrow)")
	cmd.Flags().StringVar(&page, "page", "", "Use named page as parent")
	cmd.Flags().StringVar(&text, "text", "", "Block text (single block mode)")
	cmd.Flags().StringVar(&uid, "uid", "", "Optional block uid")
	cmd.Flags().StringVar(&order, "order", "last", "Order: first|last|<int>")
	cmd.Flags().BoolVar(&open, "open", true, "Block open state")
	cmd.Flags().StringVar(&attachTo, "attach-to", "", "Find or create a section block under parent, then create content under it")
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
			uid = client.NewUID()
		}
		order := strings.TrimSpace(node.Order)
		if order == "" {
			order = "last"
		}
		open := true
		if node.Open != nil {
			open = *node.Open
		}
		*actions = append(*actions, client.CreateBlockAction(text, parent, uid, order, open))
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
			c, err := mustClient()
			if err != nil {
				return err
			}
			resp, err := c.Write(client.UpdateBlockAction(uid, text))
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
			c, err := mustClient()
			if err != nil {
				return err
			}
			resp, err := c.Write(client.DeleteBlockAction(uid))
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
			c, err := mustClient()
			if err != nil {
				return err
			}
			action := client.MoveBlockAction(uid, parentUID, order)
			resp, err := c.Write(action)
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
			c, err := mustClient()
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
			result, err := c.Q(query, nil)
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
	var today bool
	var page string

	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find a block UID by text match",
		RunE: func(cmd *cobra.Command, args []string) error {
			if text == "" {
				return errMissingFlag("text")
			}
			if today && daily != "" {
				return fmt.Errorf("--today and --daily cannot be used together")
			}
			if today {
				daily = "today"
			}
			if daily != "" && page != "" {
				return fmt.Errorf("--daily/--today and --page cannot be used together")
			}
			if daily == "" && page == "" {
				return fmt.Errorf("one of --daily, --today, or --page is required")
			}

			c, err := mustClient()
			if err != nil {
				return err
			}

			if daily != "" {
				when, err := parseDateFlexible(daily)
				if err != nil {
					return err
				}
				uid, err := c.FindBlockUID(text, "", &when)
				if err != nil {
					return err
				}
				fmt.Print(uid)
				return nil
			}

			uid, err := c.FindBlockUID(text, maybeResolveDailyTitle(page), nil)
			if err != nil {
				return err
			}
			fmt.Print(uid)
			return nil
		},
	}

	cmd.Flags().StringVar(&text, "text", "", "Block text to match (required)")
	cmd.Flags().StringVar(&daily, "daily", "", "Daily note date (YYYY-MM-DD, today, yesterday, tomorrow)")
	cmd.Flags().BoolVar(&today, "today", false, "Shorthand for --daily today")
	cmd.Flags().StringVar(&page, "page", "", "Page title (mutually exclusive with --daily/--today)")
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
