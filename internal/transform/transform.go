package transform

import "fmt"

// LookupFunc resolves a command name to an absolute path. Returns empty string if not found.
type LookupFunc func(cmd string) (string, error)

// TransformResult describes what happened to a single server.
type TransformResult struct {
	Name    string
	Action  string // "wrapped", "resolved", "unwrapped", "skipped", "already-wrapped", "native-exe"
	From    string
	To      string
	Reason  string
	Changed bool
}

// TransformServer wraps a server's command for Windows compatibility.
// It modifies and returns a copy of the server map.
func TransformServer(name string, server map[string]any, resolve bool, lookupFn LookupFunc) (map[string]any, TransformResult) {
	result := TransformResult{Name: name}
	out := copyMap(server)

	// Check for HTTP transport — skip.
	if isHTTPTransport(out) {
		result.Action = "skipped"
		result.Reason = "HTTP transport"
		return out, result
	}

	// Handle Zed nested command object: {"command": {"path": "npx", "args": [...]}}
	if cmdObj, ok := out["command"].(map[string]any); ok {
		return transformZedNested(name, out, cmdObj, resolve, lookupFn)
	}

	cmd, _ := out["command"].(string)
	if cmd == "" {
		result.Action = "skipped"
		result.Reason = "no command"
		return out, result
	}

	// Already wrapped — skip.
	if IsAlreadyWrapped(cmd) {
		result.Action = "already-wrapped"
		result.Reason = "already uses cmd.exe"
		return out, result
	}

	// Native exe — skip.
	if IsNativeExe(cmd) {
		result.Action = "native-exe"
		result.Reason = "native executable"
		return out, result
	}

	// Resolve mode: find absolute path.
	if resolve && lookupFn != nil {
		resolved, err := lookupFn(cmd)
		if err == nil && resolved != "" {
			result.Action = "resolved"
			result.From = cmd
			result.To = resolved
			result.Changed = true
			out["command"] = resolved
			return out, result
		}
		// Fall through to wrapping if resolve fails.
	}

	// Wrap with cmd.exe /c.
	result.Action = "wrapped"
	result.From = cmd
	result.To = "cmd.exe /c " + cmd
	result.Changed = true

	args := getArgs(out)
	newArgs := make([]any, 0, len(args)+2)
	newArgs = append(newArgs, "/c", cmd)
	newArgs = append(newArgs, args...)
	out["command"] = "cmd.exe"
	out["args"] = newArgs

	return out, result
}

// UnwrapServer reverses a cmd.exe /c wrapping.
func UnwrapServer(name string, server map[string]any) (map[string]any, TransformResult) {
	result := TransformResult{Name: name}
	out := copyMap(server)

	cmd, _ := out["command"].(string)
	if !IsAlreadyWrapped(cmd) {
		result.Action = "skipped"
		result.Reason = "not wrapped"
		return out, result
	}

	args := getArgs(out)
	if len(args) < 2 {
		result.Action = "skipped"
		result.Reason = "no /c argument"
		return out, result
	}

	firstArg, _ := args[0].(string)
	if firstArg != "/c" && firstArg != "/C" {
		result.Action = "skipped"
		result.Reason = "first arg is not /c"
		return out, result
	}

	originalCmd, _ := args[1].(string)
	if originalCmd == "" {
		result.Action = "skipped"
		result.Reason = "no command after /c"
		return out, result
	}

	result.Action = "unwrapped"
	result.From = "cmd.exe /c " + originalCmd
	result.To = originalCmd
	result.Changed = true

	out["command"] = originalCmd
	remaining := args[2:]
	if len(remaining) == 0 {
		delete(out, "args")
	} else {
		out["args"] = remaining
	}

	return out, result
}

// TransformAll applies TransformServer or UnwrapServer to all servers under a key.
func TransformAll(data map[string]any, key string, unwrap bool, resolve bool, lookupFn LookupFunc) (map[string]any, []TransformResult) {
	out := copyMap(data)
	servers, ok := out[key].(map[string]any)
	if !ok {
		return out, nil
	}

	transformed := make(map[string]any)
	var results []TransformResult

	for name, v := range servers {
		srv, ok := v.(map[string]any)
		if !ok {
			transformed[name] = v
			continue
		}

		var server map[string]any
		var result TransformResult
		if unwrap {
			server, result = UnwrapServer(name, srv)
		} else {
			server, result = TransformServer(name, srv, resolve, lookupFn)
		}
		transformed[name] = server
		results = append(results, result)
	}

	out[key] = transformed
	return out, results
}

// transformZedNested handles Zed's {"command": {"path": "npx", "args": [...]}} format.
func transformZedNested(name string, out map[string]any, cmdObj map[string]any, resolve bool, lookupFn LookupFunc) (map[string]any, TransformResult) {
	result := TransformResult{Name: name}

	path, _ := cmdObj["path"].(string)
	if path == "" {
		result.Action = "skipped"
		result.Reason = "no path in command object"
		return out, result
	}

	if IsAlreadyWrapped(path) {
		result.Action = "already-wrapped"
		result.Reason = "already uses cmd.exe"
		return out, result
	}

	if IsNativeExe(path) {
		result.Action = "native-exe"
		result.Reason = "native executable"
		return out, result
	}

	if resolve && lookupFn != nil {
		resolved, err := lookupFn(path)
		if err == nil && resolved != "" {
			result.Action = "resolved"
			result.From = path
			result.To = resolved
			result.Changed = true
			newCmd := copyMap(cmdObj)
			newCmd["path"] = resolved
			out["command"] = newCmd
			return out, result
		}
	}

	result.Action = "wrapped"
	result.From = path
	result.To = "cmd.exe /c " + path
	result.Changed = true

	args := getArgs(cmdObj)
	newArgs := make([]any, 0, len(args)+2)
	newArgs = append(newArgs, "/c", path)
	newArgs = append(newArgs, args...)

	newCmd := copyMap(cmdObj)
	newCmd["path"] = "cmd.exe"
	newCmd["args"] = newArgs
	out["command"] = newCmd

	return out, result
}

func isHTTPTransport(server map[string]any) bool {
	if _, ok := server["url"]; ok {
		return true
	}
	if _, ok := server["serverUrl"]; ok {
		return true
	}
	if t, ok := server["type"].(string); ok {
		if t == "sse" || t == "http" {
			return true
		}
	}
	return false
}

func getArgs(m map[string]any) []any {
	args, ok := m["args"].([]any)
	if !ok {
		return nil
	}
	return args
}

func copyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// FormatResult returns a human-readable line for a transform result.
func FormatResult(r TransformResult) string {
	switch r.Action {
	case "wrapped":
		return fmt.Sprintf("  ✓ %-20s %s → cmd.exe /c %s", r.Name, r.From, r.From)
	case "resolved":
		return fmt.Sprintf("  ✓ %-20s %s → %s", r.Name, r.From, r.To)
	case "unwrapped":
		return fmt.Sprintf("  ✓ %-20s %s → %s", r.Name, r.From, r.To)
	case "skipped":
		return fmt.Sprintf("  - %-20s (%s)", r.Name, r.Reason)
	case "already-wrapped":
		return fmt.Sprintf("  - %-20s (%s)", r.Name, r.Reason)
	case "native-exe":
		return fmt.Sprintf("  - %-20s (%s)", r.Name, r.Reason)
	default:
		return fmt.Sprintf("  ? %-20s %s", r.Name, r.Action)
	}
}
