package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"roam-cli/internal/format"
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

			client, err := mustClient()
			if err != nil {
				return err
			}
			nodes, err := client.GetJournalingByDate(when, topic)
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

	cmd.Flags().StringVar(&date, "date", "", "Date string, defaults to today")
	cmd.Flags().StringVar(&topic, "topic", "", "Topic node override")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output journal data as JSON")
	cmd.Flags().BoolVar(&asPlain, "plain", false, "Output journal data as plain text")
	return cmd
}

func parseDateFlexible(v string) (time.Time, error) {
	if strings.TrimSpace(v) == "" {
		return time.Now(), nil
	}
	layouts := []string{
		time.RFC3339,
		time.RFC1123Z,
		time.RFC1123,
		"2006-01-02",
		"2006/01/02",
		"01-02-2006",
		"01/02/2006",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, strings.TrimSpace(v)); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %s", v)
}
