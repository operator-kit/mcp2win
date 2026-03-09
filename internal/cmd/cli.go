package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/operator-kit/mcp2win/internal/color"
	"github.com/operator-kit/mcp2win/internal/config"
	"github.com/operator-kit/mcp2win/internal/provider"
	"github.com/operator-kit/mcp2win/internal/transform"
)

func runCLI(positional []string, flags Flags, cfg *config.Config, stdout, stderr io.Writer) int {
	if len(positional) == 0 {
		printUsage(stderr)
		return 1
	}

	p, consumed := provider.DetectProvider(positional)
	if p == nil {
		// Unknown provider — try generic wrapping.
		return runGenericCLI(positional, flags, cfg, stdout, stderr)
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

	// Show preview.
	preview := p.FormatOutput(parsed, transformed)
	if !flags.Quiet {
		fmt.Fprintln(stderr, color.Dim(preview))
	}

	// --dry-run: print command to stdout, don't execute.
	if flags.DryRun {
		fmt.Fprintln(stdout, preview)
		return 0
	}

	// Confirm before executing.
	if !confirmAction("Execute?", "always_exec_cli", flags, cfg, stderr) {
		return 0
	}

	executable, execArgs := p.ExecArgs(parsed, transformed)
	return ExecFn(executable, execArgs)
}

func runGenericCLI(positional []string, flags Flags, cfg *config.Config, stdout, stderr io.Writer) int {
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

	display := strings.Join(parts, " ")

	// Show preview.
	if !flags.Quiet {
		fmt.Fprintln(stderr, color.Dim(display))
	}

	// --dry-run: print command to stdout, don't execute.
	if flags.DryRun {
		fmt.Fprintln(stdout, display)
		return 0
	}

	// Confirm before executing.
	if !confirmAction("Execute?", "always_exec_cli", flags, cfg, stderr) {
		return 0
	}

	executable := parts[0]
	execArgs := parts[1:]
	return ExecFn(executable, execArgs)
}
