package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Leechael/roam-cli/internal/format"
)

func newJournalCmd() *cobra.Command {
	var date string
	var topic string
	var asJSON bool
	var asPlain bool

	cmd := &cobra.Command{
		Use:     "journal",
		Aliases: []string{"get-journaling-by-date", "journaling"},
		Short:   "Get journaling blocks from Daily Notes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if asJSON && asPlain {
				return fmt.Errorf("--json and --plain cannot be used together")
			}
			when, err := parseDateFlexible(date)
			if err != nil {
				return err
			}
			if topic == "" {
				topic = os.Getenv("TOPIC_NODE")
			}
			if topic == "" {
				topic = os.Getenv("ROAM_TOPIC_NODE")
			}

			c, err := mustClient()
			if err != nil {
				return err
			}
			nodes, err := c.GetJournalingByDate(when, topic)
			if err != nil {
				return err
			}
			if asJSON {
				return prettyPrint(nodes)
			}
			fmt.Println(format.FormatJournal(nodes, topic != ""))
			return nil
		},
	}

	cmd.Flags().StringVar(&date, "date", "", "Date (YYYY-MM-DD, today, yesterday, tomorrow); defaults to today")
	cmd.Flags().StringVar(&topic, "topic", "", "Topic node override")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output journal data as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output journal data as plain text")
	return cmd
}
