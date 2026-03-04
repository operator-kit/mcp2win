package cmd

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/operator-kit/mcp2win/internal/transform"
)

func runJSON(positional []string, flags Flags, stdin io.Reader, stdout, stderr io.Writer) int {
	var input []byte
	var err error

	if len(positional) > 0 {
		input = []byte(positional[0])
	} else {
		input, err = io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "Error reading stdin: %v\n", err)
			return 1
		}
	}

	if len(input) == 0 {
		fmt.Fprintln(stderr, "Error: no JSON input")
		return 1
	}

	var data map[string]any
	if err := json.Unmarshal(input, &data); err != nil {
		fmt.Fprintf(stderr, "Error: invalid JSON: %v\n", err)
		return 1
	}

	shape, key, err := transform.DetectShape(data)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	var result map[string]any
	var results []transform.TransformResult

	lookupFn := getLookupFn(flags.Resolve)

	switch shape {
	case transform.ShapeSingleServer:
		var r transform.TransformResult
		if flags.Unwrap {
			result, r = transform.UnwrapServer("server", data)
		} else {
			result, r = transform.TransformServer("server", data, flags.Resolve, lookupFn)
		}
		results = []transform.TransformResult{r}

	case transform.ShapeServersBlock:
		// Wrap in a temporary key to use TransformAll.
		wrapped := map[string]any{"_servers": data}
		out, rs := transform.TransformAll(wrapped, "_servers", flags.Unwrap, flags.Resolve, lookupFn)
		result = out["_servers"].(map[string]any)
		results = rs

	case transform.ShapeFullConfig:
		result, results = transform.TransformAll(data, key, flags.Unwrap, flags.Resolve, lookupFn)
	}

	if !flags.Quiet {
		printPreview(stderr, results)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, string(output))
	return 0
}
