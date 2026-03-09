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
  mcp2win [flags] <file.json>          Transform & write config file
  mcp2win [flags] '<json>'             Transform inline JSON
  echo '<json>' | mcp2win [flags]      Transform JSON from stdin
  mcp2win [flags] claude mcp add ...   Translate & run CLI command
  mcp2win config <command>             View/modify preferences
  mcp2win update                       Update to the latest version

Flags:
  --yes, -y    Skip confirmation prompt (non-interactive mode)
  --dry-run    Preview only, no action
  --quiet      Suppress preview output
  --no-backup  Skip .bak backup when writing files
  -o <path>    Write output to a different file
  --unwrap     Reverse: remove cmd.exe /c wrapping
  --resolve    Resolve commands to absolute paths via PATH/PATHEXT
  --no-color   Disable colored output
  --version    Show version
  --help       Show this help

Config:
  mcp2win config get [key]           Show preferences
  mcp2win config set <key> <value>   Set a preference
  mcp2win config path                Show config file location
  mcp2win config reset               Reset all preferences

Examples:
  mcp2win claude mcp add github-server -- npx -y @modelcontextprotocol/server-github
  mcp2win claude_desktop_config.json
  mcp2win -y claude_desktop_config.json
  mcp2win '{"command":"npx","args":["-y","@modelcontextprotocol/server-github"]}'
  mcp2win --unwrap '{"command":"cmd.exe","args":["/c","npx","-y","@pkg"]}'
`)
}
