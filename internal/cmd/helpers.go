package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/itchyny/gojq"

	"github.com/Leechael/roamresearch-skills/internal/client"
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

// maybeResolveDailyTitle checks if s looks like a date string (e.g. "2026-03-14").
// If so, it returns the Roam daily page title ("March 14th, 2026").
// Otherwise it returns s unchanged.
func maybeResolveDailyTitle(s string) string {
	t, err := parseDateFlexible(s)
	if err != nil || strings.TrimSpace(s) == "" {
		return s
	}
	return client.DailyTitle(t)
}

func parseDateFlexible(v string) (time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return time.Now(), nil
	}
	now := time.Now()
	switch strings.ToLower(v) {
	case "today":
		return now, nil
	case "yesterday":
		return now.AddDate(0, 0, -1), nil
	case "tomorrow":
		return now.AddDate(0, 0, 1), nil
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
		if t, err := time.Parse(layout, v); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %s", v)
}
