package transform

import "strings"

// Known shim commands that need cmd.exe /c wrapping on Windows.
var shimCommands = map[string]bool{
	"npx": true, "pnpx": true, "bunx": true, "uvx": true,
	"yarn": true, "poetry": true, "pipx": true,
	"tsx": true, "ts-node": true,
	"eslint": true, "prettier": true, "tsc": true,
	"vite": true, "turbo": true, "next": true, "nuxt": true,
	"playwright": true,
}

// Native executables that don't need wrapping.
var nativeCommands = map[string]bool{
	"node": true, "python": true, "python3": true,
	"deno": true, "bun": true, "cmd": true,
}

// IsKnownCommand returns true if cmd is any known command (shim or native).
func IsKnownCommand(cmd string) bool {
	cmd = normalize(cmd)
	return shimCommands[cmd] || nativeCommands[cmd]
}

// NeedsWrapping returns true if cmd is a known shim command.
func NeedsWrapping(cmd string) bool {
	cmd = normalize(cmd)
	return shimCommands[cmd]
}

// IsNativeExe returns true if cmd is a native executable that runs directly.
func IsNativeExe(cmd string) bool {
	cmd = normalize(cmd)
	return nativeCommands[cmd]
}

// IsAlreadyWrapped returns true if cmd is cmd or cmd.exe.
func IsAlreadyWrapped(cmd string) bool {
	cmd = normalize(cmd)
	return cmd == "cmd" || cmd == "cmd.exe"
}

// normalize lowercases and strips .exe suffix.
func normalize(cmd string) string {
	cmd = strings.ToLower(cmd)
	if cmd != "cmd.exe" && strings.HasSuffix(cmd, ".exe") {
		cmd = strings.TrimSuffix(cmd, ".exe")
	}
	return cmd
}
