package cmd

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	helpdocs "github.com/Leechael/roam-cli/docs/help"
)

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
		fmt.Fprintf(w, "  %-19s %s\n", t.Name, t.Desc)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Use \"roam-cli help <topic>\" for more information about a topic.")
}

func installHelpCmd(root *cobra.Command) {
	defaultHelp := root.HelpFunc()

	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelp(cmd, args)
		if cmd == root {
			fprintHelpTopicsBlock(cmd.OutOrStdout())
		}
	})

	helpCmd := &cobra.Command{
		Use:     "help [command | topic]",
		Short:   "Help about any command or topic",
		GroupID: "daily",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			if len(args) == 0 {
				defaultHelp(root, args)
				fprintHelpTopicsBlock(w)
				return nil
			}

			name := strings.ToLower(args[0])

			// Check help topics first
			content, err := readTopic(name)
			if err == nil {
				fmt.Fprint(w, content)
				return nil
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
			return fmt.Errorf("unknown help topic %q\nAvailable: %s", name, strings.Join(available, ", "))
		},
	}

	root.SetHelpCommand(helpCmd)
}
