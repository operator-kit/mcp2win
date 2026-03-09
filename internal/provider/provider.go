package provider

import (
	"strings"

	"github.com/operator-kit/mcp2win/internal/transform"
)

// ParsedCLI holds the parsed components of a CLI command.
type ParsedCLI struct {
	ServerName string
	Command    string
	Args       []string
	Scope      string
	EnvVars    map[string]string
	Extra      map[string]any // provider-specific parsed data
}

// Provider handles CLI command translation for a specific MCP client.
type Provider interface {
	Name() string
	ParseArgs(args []string) (ParsedCLI, error)
	FormatOutput(parsed ParsedCLI, transformed map[string]any) string
	ExecArgs(parsed ParsedCLI, transformed map[string]any) (string, []string)
}

// InferSeparator finds the index of the first known command in tokens,
// starting from the given offset. Used as a fallback when '--' is missing
// (e.g., consumed by npm shim). Returns -1 if no known command found.
func InferSeparator(tokens []string, startAt int) int {
	for i := startAt; i < len(tokens); i++ {
		if transform.IsKnownCommand(tokens[i]) {
			return i
		}
	}
	return -1
}

// DetectProvider returns the matching provider and how many tokens it consumed.
// Returns nil if no provider matches.
func DetectProvider(args []string) (Provider, int) {
	if len(args) == 0 {
		return nil, 0
	}

	switch strings.ToLower(args[0]) {
	case "claude":
		return &Claude{}, 1
	case "code":
		return &VSCode{}, 1
	case "qchat", "q":
		return &AmazonQ{cmd: args[0]}, 1
	case "gemini":
		return &Gemini{}, 1
	default:
		return nil, 0
	}
}
