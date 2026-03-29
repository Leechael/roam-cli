package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	helpdocs "github.com/Leechael/roamresearch-skills/docs/help"
)

type exampleCategory struct {
	Name    string
	Desc    string
	Content string
}

var categories = []exampleCategory{
	{
		Name: "read",
		Desc: "Read pages, blocks, search, query, and journal",
		Content: `## Read page or block

  roam-cli get "Page Title"
  roam-cli get "((block-uid))"
  roam-cli get "Page Title" --json

## Search blocks

  roam-cli search term1 term2 --limit 20
  roam-cli search keyword --page "Project" --ignore-case

## Datalog query

  roam-cli q '[:find ?title :where [?e :node/title ?title]]'

## Journal by date

  roam-cli journal --date 2026-02-12
  roam-cli journal --date 2026-02-12 --topic "Work Log"

## Find block UID

  roam-cli block find --text "[[📖 Daily Reading]]" --daily 2026-02-15
  roam-cli block find --text "Status" --page "Project Dashboard"

## Get block recursively

  roam-cli block get --uid <uid>`,
	},
	{
		Name: "write",
		Desc: "Save markdown, create blocks, batch operations",
		Content: `## Save to daily page (preferred for daily notes — one-shot)

  cat note.md | roam-cli save --to-daily-page
  cat note.md | roam-cli save --to-daily-page 2026-03-14

## Save markdown as page

  cat note.md | roam-cli save --title "New Page"
  roam-cli save --title "New Page" --file ./note.md
  roam-cli save --parent <uid> --file ./note.md

## Create blocks (single, nested tree, or attach-to)

  # Single block
  roam-cli block create --parent <uid> --text "hello"

  # Nested tree from JSON
  echo '{"text":"headline","children":[{"text":"child 1"},{"text":"child 2"}]}' \
    | roam-cli block create --parent <uid>
  roam-cli block create --parent <uid> --file ./tree.json

  # Attach-to: find or create section, then insert under it
  roam-cli block create --parent <page-uid> --attach-to "[[📽 Journaling]]" --text "item"
  roam-cli block create --parent <page-uid> --attach-to "[[📽 Journaling]]" --file items.json

## Batch operations (preferred for mixed action types)

  roam-cli batch run --file ./actions.json
  echo '[...]' | roam-cli batch run

## Other block operations

  roam-cli block update --uid <uid> --text "updated"
  roam-cli block move   --uid <uid> --parent <target-uid> --order last
  roam-cli block delete --uid <uid>`,
	},
	{
		Name: "workflow",
		Desc: "Common multi-step patterns for efficient Roam operations",
		Content: `## Save to today's daily page (1 call)

  cat note.md | roam-cli save --to-daily-page
  echo '# Article Summary' | roam-cli save --to-daily-page 2026-03-14

## Insert under an existing section (1 call with --attach-to)

  # Find or create "Daily Reading" section, insert under it
  roam-cli block create --parent <daily-page-uid> \
    --attach-to "[[📖 Daily Reading]]" --file article.json

## Build a project status tree (1 call, not N)

  # WRONG: sequential block create calls
  #   uid1=$(roam-cli block create --parent $PAGE --text "Project A" --plain)
  #   roam-cli block create --parent $uid1 --text "Task 1"
  #   roam-cli block create --parent $uid1 --text "Task 2"

  # RIGHT: single call with JSON tree
  echo '{"text":"Project A","children":[
    {"text":"Task 1"},
    {"text":"Task 2"},
    {"text":"Task 3","children":[{"text":"Subtask 3.1"}]}
  ]}' | roam-cli block create --parent "$PAGE"

## Search/find on a daily page (pass ISO date)

  roam-cli search --page 2026-03-14 keyword
  roam-cli block find --page 2026-03-14 --text "[[📖 Daily Reading]]"

## Save a long document as a page

  cat <<'EOF' | roam-cli save --title "Weekly Report 2026-W11"
  ## Highlights
  - Feature A shipped
  - Bug B fixed

  ## Metrics
  | Metric | Value |
  | --- | --- |
  | PRs merged | 12 |
  | Issues closed | 8 |
  EOF

## Batch: move multiple blocks + update

  echo '[
    {"action":"move-block","location":{"parent-uid":"target","order":"last"},"block":{"uid":"abc"}},
    {"action":"move-block","location":{"parent-uid":"target","order":"last"},"block":{"uid":"def"}},
    {"action":"update-block","block":{"uid":"abc","string":"moved and updated"}}
  ]' | roam-cli batch run`,
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
		Use:   "help [command | topic | category]",
		Short: "Help about any command, help topic, or example category",
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
