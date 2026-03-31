package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Leechael/roam-cli/internal/client"
)

func newMoveCmd() *cobra.Command {
	var uid string
	var parentUID string
	var title string
	var daily string
	var today bool
	var under string
	var order string
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "move",
		Short: "Move a block to a page or section",
		Long: `Move a block to a target page or section by name. The target can be:

  --parent <uid>       Direct parent block/page UID
  --title <page>       Page by title
  --today              Today's daily page
  --daily <date>       Daily page by date (YYYY-MM-DD, today, yesterday, tomorrow)

Optionally use --under to move into a specific section (find-or-create).`,
		Example: `  roam-cli move --uid <block> --title "Project X" --under "[[Tasks]]"
  roam-cli move --uid <block> --today --under "[[Archive]]"
  roam-cli move --uid <block> --daily yesterday`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			if uid == "" {
				return errMissingFlag("uid")
			}
			if today && daily != "" {
				return fmt.Errorf("--today and --daily cannot be used together")
			}
			if today {
				daily = "today"
			}

			// Count targets
			targets := 0
			if parentUID != "" {
				targets++
			}
			if title != "" {
				targets++
			}
			if daily != "" {
				targets++
			}
			if targets == 0 {
				return fmt.Errorf("one of --parent, --title, --today, or --daily is required")
			}
			if targets > 1 {
				return fmt.Errorf("--parent, --title, --today/--daily are mutually exclusive")
			}

			// --under requires a page target, not --parent
			if under != "" && parentUID != "" {
				return fmt.Errorf("--under cannot be used with --parent (use --title, --today, or --daily)")
			}

			c, err := mustClient()
			if err != nil {
				return err
			}

			// Resolve target to UID
			target := parentUID
			if daily != "" {
				when, err := parseDateFlexible(daily)
				if err != nil {
					return err
				}
				pageTitle := client.DailyTitle(when)
				pageUID, err := c.GetPageUIDByTitle(pageTitle)
				if err != nil {
					return fmt.Errorf("failed to look up daily page %q: %w", pageTitle, err)
				}
				if pageUID == "" {
					return fmt.Errorf("daily page %q does not exist", pageTitle)
				}
				target = pageUID
			}
			if title != "" {
				pageUID, err := c.GetPageUIDByTitle(title)
				if err != nil {
					return fmt.Errorf("failed to look up page %q: %w", title, err)
				}
				if pageUID == "" {
					return fmt.Errorf("page %q does not exist", title)
				}
				target = pageUID
			}

			// --under: find-or-create section
			if under != "" {
				foundUID, err := c.FindBlockUnderParent(under, target)
				if err != nil {
					return fmt.Errorf("--under lookup failed: %w", err)
				}
				if foundUID != "" {
					target = foundUID
				} else {
					newUID := client.NewUID()
					action := client.CreateBlockAction(under, target, newUID, "last", true)
					if _, err := c.Write(action); err != nil {
						return fmt.Errorf("failed to create section block: %w", err)
					}
					target = newUID
				}
			}

			action := client.MoveBlockAction(uid, target, order)
			resp, err := c.Write(action)
			if err != nil {
				return err
			}

			if asPlain {
				fmt.Println(uid)
				return nil
			}
			return prettyPrint(map[string]any{
				"uid":        uid,
				"parent_uid": target,
				"response":   resp,
			})
		},
	}

	cmd.Flags().StringVar(&uid, "uid", "", "Block UID to move (required)")
	cmd.Flags().StringVar(&parentUID, "parent", "", "Target parent block/page UID")
	cmd.Flags().StringVar(&title, "title", "", "Target page by title")
	cmd.Flags().BoolVar(&today, "today", false, "Move to today's daily page")
	cmd.Flags().StringVar(&daily, "daily", "", "Move to daily page by date (YYYY-MM-DD, today, yesterday, tomorrow)")
	cmd.Flags().StringVar(&under, "under", "", "Find-or-create a section block under the target, then move into it")
	cmd.Flags().StringVar(&order, "order", "last", "Order: first|last|<int>")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output moved block UID")

	return cmd
}
