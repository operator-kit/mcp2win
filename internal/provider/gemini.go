package provider

import (
	"fmt"
	"strings"
)

// Gemini handles `gemini mcp add NAME -- CMD ARGS...`
// Note: Gemini CLI handles Windows natively since Oct 2025.
type Gemini struct{}

func (g *Gemini) Name() string { return "gemini" }

func (g *Gemini) ParseArgs(args []string) (ParsedCLI, error) {
	if len(args) < 2 || strings.ToLower(args[0]) != "mcp" || strings.ToLower(args[1]) != "add" {
		return ParsedCLI{}, fmt.Errorf("expected: gemini mcp add NAME -- CMD ARGS...")
	}

	rest := args[2:]
	dashDashIdx := -1
	for i, a := range rest {
		if a == "--" {
			dashDashIdx = i
			break
		}
	}

	if dashDashIdx == -1 {
		return ParsedCLI{}, fmt.Errorf("expected '--' separator: gemini mcp add NAME -- CMD ARGS...")
	}

	var serverName string
	if dashDashIdx > 0 {
		serverName = rest[0]
	}

	cmdArgs := rest[dashDashIdx+1:]
	if len(cmdArgs) == 0 {
		return ParsedCLI{}, fmt.Errorf("no command after '--'")
	}

	return ParsedCLI{
		ServerName: serverName,
		Command:    cmdArgs[0],
		Args:       cmdArgs[1:],
	}, nil
}

func (g *Gemini) FormatOutput(parsed ParsedCLI, transformed map[string]any) string {
	var parts []string
	parts = append(parts, "gemini", "mcp", "add")
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

	result := strings.Join(parts, " ")
	result += fmt.Sprintf("\n\n# Note: Gemini CLI handles Windows paths natively since Oct 2025 — you may not need this")
	return result
}

func (g *Gemini) ExecArgs(parsed ParsedCLI, transformed map[string]any) (string, []string) {
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

	return "gemini", args
}
