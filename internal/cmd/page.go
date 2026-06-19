package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPageCmd() *cobra.Command {
	page := &cobra.Command{
		Use:   "page",
		Short: "Page operations",
	}

	page.AddCommand(newPageClearCmd())
	page.AddCommand(newPageDeleteCmd())

	return page
}

func newPageClearCmd() *cobra.Command {
	var daily string
	var today bool
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "clear [title]",
		Short: "Clear a page's content",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			title, err := resolvePageTarget(args, daily, today)
			if err != nil {
				return err
			}
			c, err := mustClient()
			if err != nil {
				return err
			}
			pageUID, deleted, resp, err := clearPageContent(c, title)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"title":    title,
				"page_uid": pageUID,
				"deleted":  deleted,
				"response": resp,
			}
			if asJSON {
				return prettyPrint(payload)
			}
			if asPlain {
				fmt.Println(pageUID)
				return nil
			}
			fmt.Printf("cleared page %q (%s) with %d blocks\n", title, pageUID, deleted)
			return nil
		},
	}

	cmd.Flags().StringVar(&daily, "daily", "", "Page date (YYYY-MM-DD, today, yesterday, tomorrow)")
	cmd.Flags().BoolVar(&today, "today", false, "Shorthand for --daily today")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output result as plain text")
	return cmd
}

func newPageDeleteCmd() *cobra.Command {
	var daily string
	var today bool
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:   "delete [title]",
		Short: "Delete a page",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateOutputFlags(asJSON, asPlain); err != nil {
				return err
			}
			title, err := resolvePageTarget(args, daily, today)
			if err != nil {
				return err
			}
			c, err := mustClient()
			if err != nil {
				return err
			}
			pageUID, resp, err := deletePageByTitle(c, title)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"title":    title,
				"page_uid": pageUID,
				"response": resp,
			}
			if asJSON {
				return prettyPrint(payload)
			}
			if asPlain {
				fmt.Println(pageUID)
				return nil
			}
			fmt.Printf("deleted page %q (%s)\n", title, pageUID)
			return nil
		},
	}

	cmd.Flags().StringVar(&daily, "daily", "", "Page date (YYYY-MM-DD, today, yesterday, tomorrow)")
	cmd.Flags().BoolVar(&today, "today", false, "Shorthand for --daily today")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output result as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output result as plain text")
	return cmd
}
