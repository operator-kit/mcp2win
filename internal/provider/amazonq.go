package provider

import (
	"fmt"
	"strings"
)

// AmazonQ handles `qchat mcp add -- CMD ARGS...` and `q mcp add -- CMD ARGS...`
type AmazonQ struct {
	cmd string // "qchat" or "q"
}

func (a *AmazonQ) Name() string { return a.cmd }

func (a *AmazonQ) ParseArgs(args []string) (ParsedCLI, error) {
	// args: ["mcp", "add", "--", "CMD", "ARGS..."]
	// or: ["mcp", "add", "NAME", "--", "CMD", "ARGS..."]
	if len(args) < 2 || strings.ToLower(args[0]) != "mcp" || strings.ToLower(args[1]) != "add" {
		return ParsedCLI{}, fmt.Errorf("expected: %s mcp add [NAME] -- CMD ARGS...", a.cmd)
	}

	rest := args[2:]
	dashDashIdx := -1
	for i, a := range rest {
		if a == "--" {
			dashDashIdx = i
			break
		}
	}

	var serverName string
	var cmdArgs []string

	if dashDashIdx >= 0 {
		if dashDashIdx > 0 {
			serverName = rest[dashDashIdx-1]
		}
		cmdArgs = rest[dashDashIdx+1:]
		if len(cmdArgs) == 0 {
			return ParsedCLI{}, fmt.Errorf("no command after '--'")
		}
	} else {
		// No separator — heuristic fallback.
		// Try index >= 1 first (prefer leaving room for server name), then 0.
		cmdIdx := InferSeparator(rest, 1)
		if cmdIdx < 0 {
			cmdIdx = InferSeparator(rest, 0)
		}
		if cmdIdx < 0 {
			return ParsedCLI{}, fmt.Errorf("expected '--' separator: %s mcp add [NAME] -- CMD ARGS...", a.cmd)
		}
		if cmdIdx > 0 {
			serverName = rest[cmdIdx-1]
		}
		cmdArgs = rest[cmdIdx:]
	}

	return ParsedCLI{
		ServerName: serverName,
		Command:    cmdArgs[0],
		Args:       cmdArgs[1:],
	}, nil
}

func (a *AmazonQ) FormatOutput(parsed ParsedCLI, transformed map[string]any) string {
	var parts []string
	parts = append(parts, a.cmd, "mcp", "add")
	if parsed.ServerName != "" {
		parts = append(parts, parsed.ServerName)
	}
	parts = append(parts, "--")

	cmd, _ := transformed["command"].(string)
	parts = append(parts, cmd)

	if args, ok := transformed["args"].([]any); ok {
		for _, arg := range args {
			if s, ok := arg.(string); ok {
				parts = append(parts, s)
			}
		}
	}

	return strings.Join(parts, " ")
}

func (a *AmazonQ) ExecArgs(parsed ParsedCLI, transformed map[string]any) (string, []string) {
	var args []string
	args = append(args, "mcp", "add")
	if parsed.ServerName != "" {
		args = append(args, parsed.ServerName)
	}
	args = append(args, "--")

	cmd, _ := transformed["command"].(string)
	args = append(args, cmd)
	if tArgs, ok := transformed["args"].([]any); ok {
		for _, arg := range tArgs {
			if s, ok := arg.(string); ok {
				args = append(args, s)
			}
		}
	}

	return a.cmd, args
}
