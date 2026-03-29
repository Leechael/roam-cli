package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/itchyny/gojq"
)

// Formatter handles output in JSON, plain, or human-readable modes.
type Formatter struct {
	json   bool
	plain  bool
	jqExpr string
}

// New creates a Formatter from the given flags.
func New(asJSON, asPlain bool, jqExpr string) *Formatter {
	return &Formatter{json: asJSON, plain: asPlain, jqExpr: jqExpr}
}

// Validate checks that the flag combination is valid.
func (f *Formatter) Validate() error {
	if f.json && f.plain {
		return fmt.Errorf("--json and --plain cannot be used together")
	}
	if f.jqExpr != "" && !f.json {
		return fmt.Errorf("--jq requires --json")
	}
	return nil
}

// IsJSON returns true if JSON output mode is enabled.
func (f *Formatter) IsJSON() bool { return f.json }

// IsPlain returns true if plain output mode is enabled.
func (f *Formatter) IsPlain() bool { return f.plain }

// Print writes data to w in the appropriate format.
// JSON mode: pretty-printed JSON (with optional jq filtering).
// Plain and default modes return nil — the caller handles output directly.
func (f *Formatter) Print(w io.Writer, data any) error {
	if f.jqExpr != "" {
		filtered, err := applyJQ(data, f.jqExpr)
		if err != nil {
			return err
		}
		return printJSON(w, filtered)
	}
	if f.json {
		return printJSON(w, data)
	}
	return nil
}

// PrintJSON writes data as pretty-printed JSON to w.
func (f *Formatter) PrintJSON(w io.Writer, data any) error {
	return printJSON(w, data)
}

// Hint writes a hint message to stderr.
func (f *Formatter) Hint(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

// PrintMessage writes a plain message to w.
func (f *Formatter) PrintMessage(w io.Writer, msg string) {
	fmt.Fprintln(w, msg)
}

func printJSON(w io.Writer, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(b))
	return nil
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
	var results []any
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
