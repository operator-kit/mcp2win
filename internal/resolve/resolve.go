package resolve

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveCommand finds the absolute path of a command by searching PATH dirs with PATHEXT extensions.
// If pathDirs or pathExt are nil, they are read from environment variables.
func ResolveCommand(cmd string, pathDirs []string, pathExt []string) (string, error) {
	// Already absolute — return as-is.
	// Check both native and Windows-style paths (for cross-platform correctness).
	if filepath.IsAbs(cmd) || isWindowsAbs(cmd) {
		return cmd, nil
	}

	if pathDirs == nil {
		pathDirs = filepath.SplitList(os.Getenv("PATH"))
	}
	if pathExt == nil {
		pathExt = splitPathExt(os.Getenv("PATHEXT"))
	}

	// If no PATHEXT (non-Windows), try the command directly.
	if len(pathExt) == 0 {
		pathExt = []string{""}
	}

	for _, dir := range pathDirs {
		for _, ext := range pathExt {
			candidate := filepath.Join(dir, cmd+ext)
			if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
				return candidate, nil
			}
		}
	}

	return "", fmt.Errorf("command not found: %s", cmd)
}

// isWindowsAbs detects Windows absolute paths (e.g. C:\...) even on Linux.
func isWindowsAbs(path string) bool {
	if len(path) < 3 {
		return false
	}
	return path[1] == ':' && (path[2] == '\\' || path[2] == '/')
}

func splitPathExt(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(strings.ToLower(s), ";")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
