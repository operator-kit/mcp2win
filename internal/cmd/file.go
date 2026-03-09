package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/operator-kit/mcp2win/internal/color"
	"github.com/operator-kit/mcp2win/internal/config"
	"github.com/operator-kit/mcp2win/internal/resolve"
	"github.com/operator-kit/mcp2win/internal/transform"
)

func runFile(positional []string, flags Flags, cfg *config.Config, stdout, stderr io.Writer) int {
	if len(positional) == 0 {
		fmt.Fprintln(stderr, "Error: no file specified")
		return 1
	}

	filePath := positional[0]
	input, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	var data map[string]any
	if err := json.Unmarshal(input, &data); err != nil {
		fmt.Fprintf(stderr, "Error: invalid JSON in %s: %v\n", filePath, err)
		return 1
	}

	shape, key, err := transform.DetectShape(data)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	lookupFn := getLookupFn(flags.Resolve)

	var result map[string]any
	var results []transform.TransformResult

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
		wrapped := map[string]any{"_servers": data}
		out, rs := transform.TransformAll(wrapped, "_servers", flags.Unwrap, flags.Resolve, lookupFn)
		result = out["_servers"].(map[string]any)
		results = rs

	case transform.ShapeFullConfig:
		result, results = transform.TransformAll(data, key, flags.Unwrap, flags.Resolve, lookupFn)
	}

	// Show preview.
	if !flags.Quiet {
		printPreview(stderr, results)
	}

	// --dry-run: preview only, no write.
	if flags.DryRun {
		return 0
	}

	// Check if there are any changes to write.
	anyChanged := false
	for _, r := range results {
		if r.Changed {
			anyChanged = true
			break
		}
	}
	if !anyChanged {
		return 0
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	outPath := filePath
	if flags.Output != "" {
		outPath = flags.Output
	}

	// Confirm before writing.
	prompt := fmt.Sprintf("Write changes to %s?", outPath)
	if !confirmAction(prompt, "always_write_file", flags, cfg, stderr) {
		// Still output JSON to stdout so it's not lost.
		fmt.Fprintln(stdout, string(output))
		return 0
	}

	// Create backup unless --no-backup or writing to a different file.
	if !flags.NoBackup && flags.Output == "" {
		bakPath := nextBackupPath(filePath)
		if err := os.WriteFile(bakPath, input, 0644); err != nil {
			fmt.Fprintf(stderr, "Error creating backup: %v\n", err)
			return 1
		}
		if !flags.Quiet {
			fmt.Fprintf(stderr, "%s %s\n", color.Dim("Backup:"), bakPath)
		}
	}

	if err := os.WriteFile(outPath, append(output, '\n'), 0644); err != nil {
		fmt.Fprintf(stderr, "Error writing file: %v\n", err)
		return 1
	}
	if !flags.Quiet {
		fmt.Fprintf(stderr, "%s %s\n", color.Green("Written:"), outPath)
	}
	return 0
}

func printPreview(w io.Writer, results []transform.TransformResult) {
	if len(results) == 0 {
		return
	}

	changed := 0
	skipped := 0
	for _, r := range results {
		fmt.Fprintln(w, transform.FormatResult(r))
		if r.Changed {
			changed++
		} else {
			skipped++
		}
	}

	summary := fmt.Sprintf("Changes: %d", changed)
	if skipped > 0 {
		summary += fmt.Sprintf("  |  Skipped: %d", skipped)
	}
	fmt.Fprintln(w, summary)
}

// nextBackupPath returns the next available .bak path, incrementing
// the suffix if .bak already exists (e.g. .bak2, .bak3, ...).
func nextBackupPath(filePath string) string {
	bakPath := filePath + ".bak"
	if _, err := os.Stat(bakPath); os.IsNotExist(err) {
		return bakPath
	}
	for i := 2; ; i++ {
		candidate := filePath + ".bak" + strconv.Itoa(i)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func getLookupFn(doResolve bool) transform.LookupFunc {
	if !doResolve {
		return nil
	}
	return func(cmd string) (string, error) {
		return resolve.ResolveCommand(cmd, nil, nil)
	}
}
