package main

import (
	"fmt"
	"os"

	"github.com/Leechael/roam-cli/internal/cmd"
)

func main() {
	root := cmd.NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(cmd.ExitCode(err))
	}
}
