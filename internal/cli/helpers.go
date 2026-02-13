package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/itchyny/gojq"
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

func applyJQ(input any, expr string) (any, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return input, nil
	}
	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, err
	}
	iter := query.Run(input)
	results := []any{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, err
		}
		results = append(results, v)
	}
	if len(results) == 0 {
		return nil, nil
	}
	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}

func validateOutputFlags(asJSON, asPlain bool) error {
	if asJSON && asPlain {
		return fmt.Errorf("--json and --plain cannot be used together")
	}
	return nil
}
