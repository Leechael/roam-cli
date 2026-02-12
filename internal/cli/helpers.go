package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func prettyPrint(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func readAllFromFileOrStdin(file string, useStdin bool) (string, error) {
	if useStdin {
		b, err := io.ReadAll(os.Stdin)
		return string(b), err
	}
	if file == "" {
		return "", fmt.Errorf("provide --file or --stdin")
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readQueryArgOrStdin(queryArg string) (string, error) {
	if strings.TrimSpace(queryArg) != "" {
		return queryArg, nil
	}
	b, err := io.ReadAll(bufio.NewReader(os.Stdin))
	if err != nil {
		return "", err
	}
	q := strings.TrimSpace(string(b))
	if q == "" {
		return "", fmt.Errorf("empty query")
	}
	return q, nil
}
