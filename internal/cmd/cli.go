package cmd

import (
	"fmt"
	"io"

	"github.com/operator-kit/mcp2win/internal/color"
	"github.com/operator-kit/mcp2win/internal/provider"
	"github.com/operator-kit/mcp2win/internal/transform"
)

func runCLI(positional []string, flags Flags, stdout, stderr io.Writer) int {
	if len(positional) == 0 {
		printUsage(stderr)
		return 1
	}

	p, consumed := provider.DetectProvider(positional)
	if p == nil {
		// Unknown provider — try generic wrapping.
		return runGenericCLI(positional, flags, stdout, stderr)
	}

	remaining := positional[consumed:]
	parsed, err := p.ParseArgs(remaining)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	// Check if command needs wrapping.
	if transform.IsAlreadyWrapped(parsed.Command) {
		fmt.Fprintf(stderr, "%s command is already wrapped with cmd.exe\n", color.Yellow("Note:"))
		return 0
	}
	if transform.IsNativeExe(parsed.Command) {
		fmt.Fprintf(stderr, "%s %s is a native executable, no wrapping needed\n", color.Yellow("Note:"), parsed.Command)
		return 0
	}

	// Build server map for transformation.
	server := map[string]any{
		"command": parsed.Command,
	}
	if len(parsed.Args) > 0 {
		args := make([]any, len(parsed.Args))
		for i, a := range parsed.Args {
			args[i] = a
		}
		server["args"] = args
	}

	lookupFn := getLookupFn(flags.Resolve)
	transformed, result := transform.TransformServer(parsed.ServerName, server, flags.Resolve, lookupFn)

	if !result.Changed {
		fmt.Fprintf(stderr, "%s no changes needed (%s)\n", color.Yellow("Note:"), result.Reason)
		return 0
	}

	output := p.FormatOutput(parsed, transformed)
	fmt.Fprintln(stdout, output)
	return 0
}

func runGenericCLI(positional []string, flags Flags, stdout, stderr io.Writer) int {
	// Find -- separator.
	dashIdx := -1
	for i, a := range positional {
		if a == "--" {
			dashIdx = i
			break
		}
	}

	if dashIdx == -1 || dashIdx+1 >= len(positional) {
		fmt.Fprintf(stderr, "Error: unknown provider %q — expected '--' separator before command\n", positional[0])
		fmt.Fprintln(stderr, "Known providers: claude, code, qchat, q, gemini")
		return 1
	}

	cmdArgs := positional[dashIdx+1:]
	cmd := cmdArgs[0]

	if transform.IsAlreadyWrapped(cmd) {
		fmt.Fprintf(stderr, "%s command is already wrapped with cmd.exe\n", color.Yellow("Note:"))
		return 0
	}
	if transform.IsNativeExe(cmd) {
		fmt.Fprintf(stderr, "%s %s is a native executable, no wrapping needed\n", color.Yellow("Note:"), cmd)
		return 0
	}

	// Generic wrapping: prefix with cmd.exe /c.
	prefix := positional[:dashIdx]
	var parts []string
	parts = append(parts, prefix...)
	parts = append(parts, "--", "cmd.exe", "/c")
	parts = append(parts, cmdArgs...)

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " "
		}
		result += p
	}

	fmt.Fprintln(stdout, result)
	return 0
}
