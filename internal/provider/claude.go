package provider

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Claude handles `claude mcp add [--scope X] [--env K=V]... NAME -- CMD ARGS...`
// Rewrites to `claude mcp add-json NAME '{"command":"cmd.exe","args":["/c",...]}'`
type Claude struct{}

func (c *Claude) Name() string { return "claude" }

func (c *Claude) ParseArgs(args []string) (ParsedCLI, error) {
	// args: ["mcp", "add", ...flags..., "NAME", "--", "CMD", "ARGS..."]
	if len(args) < 2 || strings.ToLower(args[0]) != "mcp" || strings.ToLower(args[1]) != "add" {
		return ParsedCLI{}, fmt.Errorf("expected: claude mcp add [flags] NAME -- CMD ARGS...")
	}

	rest := args[2:] // after "mcp add"

	var scope string
	envVars := make(map[string]string)
	var serverName string
	var cmdArgs []string

	// Find -- separator.
	dashDashIdx := -1
	for i, a := range rest {
		if a == "--" {
			dashDashIdx = i
			break
		}
	}

	if dashDashIdx >= 0 {
		// Explicit separator — parse flags from tokens before --.
		before := rest[:dashDashIdx]
		if dashDashIdx+1 < len(rest) {
			cmdArgs = rest[dashDashIdx+1:]
		}
		if len(cmdArgs) == 0 {
			return ParsedCLI{}, fmt.Errorf("no command after '--'")
		}

		var nonFlags []string
		for i := 0; i < len(before); i++ {
			a := before[i]
			switch {
			case a == "--scope" || a == "-s":
				if i+1 < len(before) {
					i++
					scope = before[i]
				}
			case a == "--env" || a == "-e":
				if i+1 < len(before) {
					i++
					parts := strings.SplitN(before[i], "=", 2)
					if len(parts) == 2 {
						envVars[parts[0]] = parts[1]
					}
				}
			default:
				nonFlags = append(nonFlags, a)
			}
		}
		if len(nonFlags) == 0 {
			return ParsedCLI{}, fmt.Errorf("no server name specified")
		}
		serverName = nonFlags[len(nonFlags)-1]
	} else {
		// No separator — heuristic: find first known command in non-flag tokens.
		// This handles cases where '--' is consumed by npm shim layers.
		var nonFlags []string
		for i := 0; i < len(rest); i++ {
			a := rest[i]
			switch {
			case a == "--scope" || a == "-s":
				if i+1 < len(rest) {
					i++
					scope = rest[i]
				}
			case a == "--env" || a == "-e":
				if i+1 < len(rest) {
					i++
					parts := strings.SplitN(rest[i], "=", 2)
					if len(parts) == 2 {
						envVars[parts[0]] = parts[1]
					}
				}
			default:
				nonFlags = append(nonFlags, a)
			}
		}

		cmdIdx := InferSeparator(nonFlags, 1) // skip index 0 (server name required)
		if cmdIdx < 1 {
			return ParsedCLI{}, fmt.Errorf("expected '--' separator: claude mcp add [flags] NAME -- CMD ARGS...")
		}
		serverName = nonFlags[cmdIdx-1]
		cmdArgs = nonFlags[cmdIdx:]
	}

	return ParsedCLI{
		ServerName: serverName,
		Command:    cmdArgs[0],
		Args:       cmdArgs[1:],
		Scope:      scope,
		EnvVars:    envVars,
	}, nil
}

func (c *Claude) FormatOutput(parsed ParsedCLI, transformed map[string]any) string {
	// Build the add-json command.
	// claude mcp add-json [--scope X] NAME '{"command":"cmd.exe","args":["/c",...],"env":{...}}'
	var parts []string
	parts = append(parts, "claude", "mcp", "add-json")

	if parsed.Scope != "" {
		parts = append(parts, "--scope", parsed.Scope)
	}

	parts = append(parts, parsed.ServerName)

	// Build server JSON including env if present.
	serverDef := make(map[string]any)
	for k, v := range transformed {
		serverDef[k] = v
	}
	if len(parsed.EnvVars) > 0 {
		serverDef["env"] = parsed.EnvVars
	}

	jsonBytes, _ := json.Marshal(serverDef)
	parts = append(parts, "'"+string(jsonBytes)+"'")

	result := strings.Join(parts, " ")
	result += "\n\n# Note: Using add-json instead of add to work around Claude's /c flag mangling bug"
	return result
}

func (c *Claude) ExecArgs(parsed ParsedCLI, transformed map[string]any) (string, []string) {
	var args []string
	args = append(args, "mcp", "add-json")
	if parsed.Scope != "" {
		args = append(args, "--scope", parsed.Scope)
	}
	args = append(args, parsed.ServerName)

	serverDef := make(map[string]any)
	for k, v := range transformed {
		serverDef[k] = v
	}
	if len(parsed.EnvVars) > 0 {
		serverDef["env"] = parsed.EnvVars
	}
	jsonBytes, _ := json.Marshal(serverDef)
	args = append(args, string(jsonBytes))

	return "claude", args
}
