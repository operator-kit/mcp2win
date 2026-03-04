package cmd

import (
	"fmt"
	"io"
)

func printVersion(w io.Writer) {
	fmt.Fprintf(w, "mcp2win %s (%s, %s)\n", appVersion, appCommit, appDate)
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `mcp2win — convert MCP server configs to Windows format

Usage:
  mcp2win [flags] <file.json>          Transform a config file (Mode 3)
  mcp2win [flags] '<json>'             Transform inline JSON (Mode 2)
  echo '<json>' | mcp2win [flags]      Transform JSON from stdin (Mode 2)
  mcp2win [flags] claude mcp add ...   Translate CLI command (Mode 1)

Flags:
  --write        Write changes back to the file (creates .bak backup)
  --no-backup    Skip .bak backup when using --write
  -o <path>      Write output to a different file
  --dry-run      Show preview only, no JSON output
  --quiet        Suppress preview, output only JSON
  --unwrap       Reverse: remove cmd.exe /c wrapping
  --resolve      Resolve commands to absolute paths via PATH/PATHEXT
  --no-color     Disable colored output
  --version      Show version
  --help         Show this help

Examples:
  mcp2win claude_desktop_config.json
  mcp2win --write claude_desktop_config.json
  mcp2win '{"command":"npx","args":["-y","@modelcontextprotocol/server-github"]}'
  mcp2win claude mcp add github-server -- npx -y @modelcontextprotocol/server-github
  mcp2win --unwrap '{"command":"cmd.exe","args":["/c","npx","-y","@pkg"]}'
`)
}
