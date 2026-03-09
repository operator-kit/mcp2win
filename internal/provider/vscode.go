package provider

import (
	"encoding/json"
	"fmt"
	"strings"
)

// VSCode handles `code --add-mcp '{"name":"...", "command":"..."}'`
type VSCode struct{}

func (v *VSCode) Name() string { return "vscode" }

func (v *VSCode) ParseArgs(args []string) (ParsedCLI, error) {
	// args: ["--add-mcp", '{"name":"...","command":"..."}']
	addMCPIdx := -1
	for i, a := range args {
		if a == "--add-mcp" {
			addMCPIdx = i
			break
		}
	}

	if addMCPIdx == -1 || addMCPIdx+1 >= len(args) {
		return ParsedCLI{}, fmt.Errorf("expected: code --add-mcp '<json>'")
	}

	jsonStr := args[addMCPIdx+1]
	// Strip surrounding quotes if present.
	jsonStr = strings.Trim(jsonStr, "'\"")

	var data map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return ParsedCLI{}, fmt.Errorf("invalid JSON in --add-mcp: %v", err)
	}

	name, _ := data["name"].(string)
	cmd, _ := data["command"].(string)

	var cmdArgs []string
	if rawArgs, ok := data["args"].([]any); ok {
		for _, a := range rawArgs {
			if s, ok := a.(string); ok {
				cmdArgs = append(cmdArgs, s)
			}
		}
	}

	return ParsedCLI{
		ServerName: name,
		Command:    cmd,
		Args:       cmdArgs,
		Extra:      data,
	}, nil
}

func (v *VSCode) FormatOutput(parsed ParsedCLI, transformed map[string]any) string {
	// Merge transformed command/args back into original data.
	data := make(map[string]any)
	if parsed.Extra != nil {
		for k, v := range parsed.Extra {
			data[k] = v
		}
	}
	data["command"] = transformed["command"]
	if args, ok := transformed["args"]; ok {
		data["args"] = args
	}

	jsonBytes, _ := json.Marshal(data)
	return fmt.Sprintf("code --add-mcp '%s'", string(jsonBytes))
}

func (v *VSCode) ExecArgs(parsed ParsedCLI, transformed map[string]any) (string, []string) {
	data := make(map[string]any)
	if parsed.Extra != nil {
		for k, val := range parsed.Extra {
			data[k] = val
		}
	}
	data["command"] = transformed["command"]
	if args, ok := transformed["args"]; ok {
		data["args"] = args
	}
	jsonBytes, _ := json.Marshal(data)
	return "code", []string{"--add-mcp", string(jsonBytes)}
}
