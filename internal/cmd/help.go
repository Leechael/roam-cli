package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	helpdocs "github.com/Leechael/roam-cli/docs/help"
)

type exampleCategory struct {
	Name    string
	Desc    string
	Content string
}

var categories = []exampleCategory{
	{
		Name: "read",
		Desc: "Read pages, blocks, daily notes, and search",
		Content: `## Read today's daily page

  roam-cli get --today
  roam-cli get --today --json
  roam-cli get --daily yesterday

## Read a page or block

  roam-cli get "Page Title"
  roam-cli get "((block-uid))"
  roam-cli get "Page Title" --json

## Journal by date

  roam-cli journal --date today
  roam-cli journal --date yesterday --topic "Work Log"

## Search blocks

  roam-cli search term1 term2 --limit 20
  roam-cli search keyword --page "Project" --ignore-case

## Search aggregated by page (default)

  roam-cli search "meeting" "action item" --limit 10
  roam-cli search "TODO" --daily-topic "[[TODO]]" --json

## Search individual blocks

  roam-cli search keyword --type block --limit 20
  roam-cli search keyword --type block --page "Project" -i

## Datalog query

  roam-cli q '[:find ?title :where [?e :node/title ?title]]'

## Find block UID (low-level)

  roam-cli block find --text "[[📖 Daily Reading]]" --today
  roam-cli block find --text "Status" --page "Project Dashboard"`,
	},
	{
		Name: "write",
		Desc: "Save GFM markdown and create content",
		Content: `## Save GFM markdown to today's daily page (recommended)

  printf '- journal entry' | roam-cli save --today
  printf '- entry' | roam-cli save --today --under '[[📽 Journaling]]'
  cat highlights.md | roam-cli save --today --under '[[📖 Daily Reading]]'

## Create TODOs

  printf '- {{[[TODO]]}} Review PR\n- {{[[TODO]]}} Call dentist' \
    | roam-cli save --today --under '[[TODO]]'

## Save to a named page

  cat note.md | roam-cli save --title "New Page"
  roam-cli save --title "Project X" --under '[[Tasks]]' --file ./tasks.md

## Get UID back for follow-up

  UID=$(printf '- item' | roam-cli save --today --under '[[Inbox]]' --plain)
  printf '- detail' | roam-cli save --parent "$UID"

## Low-level: block create (JSON input, explicit UIDs)

  roam-cli block create --parent <uid> --text "hello"
  echo '{"text":"Root","children":[{"text":"Child"}]}' \
    | roam-cli block create --parent <uid>
  roam-cli block create --parent <uid> --attach-to "[[Section]]" --file tree.json

## Low-level: batch operations

  roam-cli batch run --file ./actions.json
  echo '[...]' | roam-cli batch run

## Other block operations

  roam-cli block update --uid <uid> --text "updated"
  roam-cli block delete --uid <uid>`,
	},
	{
		Name: "workflow",
		Desc: "Common daily workflows and multi-step patterns",
		Content: `## Daily capture (most common operation)

  # Quick note to today's journal section
  printf '- Had a great idea about X' | roam-cli save --today --under '[[📽 Journaling]]'

  # Meeting notes under today's page
  cat meeting.md | roam-cli save --today --under '[[Meeting Notes]]'

## Morning review

  # Check today's daily page
  roam-cli get --today

  # Check yesterday's journal
  roam-cli journal --date yesterday

## Organize: move blocks to project pages

  # Move a block from daily page to a project
  roam-cli move --uid <block> --title "Project X" --under '[[Tasks]]'

  # Move to today's archive section
  roam-cli move --uid <block> --today --under '[[Archive]]'

## Save and follow up (composing commands)

  # Save content, get UID, add more under it
  UID=$(printf '- headline' | roam-cli save --today --under '[[📽 Journaling]]' --plain)
  printf '- detail 1\n- detail 2' | roam-cli save --parent "$UID"

## Build a tree in one call (not N calls)

  # WRONG: sequential block create calls
  #   uid1=$(roam-cli block create --parent $PAGE --text "Project A" --plain)
  #   roam-cli block create --parent $uid1 --text "Task 1"

  # RIGHT: single call with JSON tree
  echo '{"text":"Project A","children":[
    {"text":"Task 1"},
    {"text":"Task 2"}
  ]}' | roam-cli block create --parent "$PAGE"

## Batch: move + update multiple blocks

  echo '[
    {"action":"move-block","location":{"parent-uid":"target","order":"last"},"block":{"uid":"abc"}},
    {"action":"update-block","block":{"uid":"abc","string":"moved and updated"}}
  ]' | roam-cli batch run

## Prefer printf piping over --text flags

  # Shell-safe: printf preserves [[ ]] and emoji
  printf '- [[📽 Journaling]] entry with [[links]]' | roam-cli save --today`,
	},
}

type helpTopic struct {
	Name string
	Desc string
}

func listTopics() []helpTopic {
	entries, err := helpdocs.Topics.ReadDir("topics")
	if err != nil {
		return nil
	}
	var topics []helpTopic
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		desc := topicDescription(e.Name())
		topics = append(topics, helpTopic{Name: name, Desc: desc})
	}
	return topics
}

func topicDescription(filename string) string {
	data, err := helpdocs.Topics.ReadFile("topics/" + filename)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}

func readTopic(name string) (string, error) {
	data, err := helpdocs.Topics.ReadFile("topics/" + name + ".md")
	if err != nil {
		return "", fmt.Errorf("unknown help topic %q", name)
	}
	return string(data), nil
}

func fprintHelpTopicsBlock(w io.Writer) {
	topics := listTopics()
	if len(topics) == 0 {
		return
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "HELP TOPICS")
	for _, t := range topics {
		fmt.Fprintf(w, "  %-14s %s\n", t.Name, t.Desc)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Use \"roam-cli help <topic>\" for more information about a topic.")
}

func fprintCategoryFooter(w io.Writer) {
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Example categories (roam-cli help <category>):")
	for _, cat := range categories {
		fmt.Fprintf(w, "  %-12s %s\n", cat.Name, cat.Desc)
	}
	fmt.Fprintf(w, "  %-12s %s\n", "all", "Show all categories")
}

func installHelpCmd(root *cobra.Command) {
	defaultHelp := root.HelpFunc()

	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelp(cmd, args)
		if cmd == root {
			w := cmd.OutOrStdout()
			fprintHelpTopicsBlock(w)
			fprintCategoryFooter(w)
		}
	})

	helpCmd := &cobra.Command{
		Use:     "help [command | topic | category]",
		Short:   "Help about any command, help topic, or example category",
		GroupID: "daily",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			if len(args) == 0 {
				defaultHelp(root, args)
				fprintHelpTopicsBlock(w)
				fprintCategoryFooter(w)
				return nil
			}

			name := strings.ToLower(args[0])

			// Check help topics first
			content, err := readTopic(name)
			if err == nil {
				fmt.Fprint(w, content)
				return nil
			}

			// Check categories
			if name == "all" {
				return fprintAll(w)
			}
			for _, cat := range categories {
				if cat.Name == name {
					return fprintCategory(w, cat)
				}
			}

			// Fall back to cobra subcommand help
			sub, _, findErr := root.Find(args)
			if findErr == nil && sub != root {
				return sub.Help()
			}

			var available []string
			for _, t := range listTopics() {
				available = append(available, t.Name)
			}
			for _, cat := range categories {
				available = append(available, cat.Name)
			}
			return fmt.Errorf("unknown help topic %q\nAvailable: %s", name, strings.Join(available, ", "))
		},
	}

	root.SetHelpCommand(helpCmd)
}

func fprintCategory(w io.Writer, cat exampleCategory) error {
	fmt.Fprintf(w, "=== %s — %s ===\n\n", cat.Name, cat.Desc)
	fmt.Fprintln(w, cat.Content)
	fmt.Fprintln(w)
	return nil
}

func fprintAll(w io.Writer) error {
	for i, cat := range categories {
		if err := fprintCategory(w, cat); err != nil {
			return err
		}
		if i < len(categories)-1 {
			fmt.Fprintln(w)
		}
	}
	return nil
}
