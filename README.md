# mcp2win

Finally, MCP servers that just work on Windows.

Every MCP server README assumes macOS or Linux. You copy the config, paste it in, and nothing happens — because `npx`, `uvx`, and friends are `.cmd` batch shims on Windows, not real executables. `mcp2win` fixes this by wrapping commands with `cmd.exe /c` so your MCP servers actually start.

Works with **Claude Code/Desktop**, **VS Code**, **Cursor**, **Zed**, **Amazon Q**, and **Gemini CLI**

Transform config files, translate platform CLI commands, or pipe JSON directly - we've got you covered.

## Quick start

No install needed — just run with `npx`:

```bash
# Fix a config file
npx @operatorkit/mcp2win --write claude_desktop_config.json

# Translate a CLI command
npx @operatorkit/mcp2win claude mcp add github-server -- npx -y @modelcontextprotocol/server-github

# Convert inline JSON
npx @operatorkit/mcp2win '{"command":"npx","args":["-y","@pkg"]}'
```

## Install

If you use it often, install globally:

```bash
npm i -g @operatorkit/mcp2win

# or with go
go install github.com/operator-kit/mcp2win/cmd/mcp2win@latest

# or download a binary from GitHub Releases
```

## Usage

### CLI command translation

Paste any provider's `mcp add` command and get the Windows version:

```bash
# Claude
mcp2win claude mcp add github-server -- npx -y @modelcontextprotocol/server-github
# → claude mcp add-json github-server '{"command":"cmd.exe","args":["/c","npx","-y","@modelcontextprotocol/server-github"]}'

# VS Code
mcp2win code --add-mcp '{"name":"my-server","command":"npx","args":["-y","@pkg"]}'

# Amazon Q
mcp2win qchat mcp add -- npx -y @pkg

# Gemini
mcp2win gemini mcp add fetch-server -- npx -y @pkg
```

### JSON conversion

Transform inline JSON or pipe from stdin:

```bash
# Inline
mcp2win '{"command":"npx","args":["-y","@modelcontextprotocol/server-github"]}'

# Stdin
cat claude_desktop_config.json | mcp2win

# Full config
mcp2win '{"mcpServers":{"s1":{"command":"npx","args":["-y","@pkg"]}}}'
```

### File transformation

```bash
# Preview changes (default)
mcp2win claude_desktop_config.json

# Write changes back (creates .bak backup)
mcp2win --write claude_desktop_config.json

# Write without backup
mcp2win --write --no-backup claude_desktop_config.json

# Write to a different file
mcp2win -o windows_config.json claude_desktop_config.json

# Dry run (preview only, no JSON output)
mcp2win --dry-run claude_desktop_config.json
```

## Flags

| Flag | Description |
|---|---|
| `--write` | Write changes back to the file (creates `.bak` backup) |
| `--no-backup` | Skip `.bak` creation when using `--write` |
| `-o <path>` | Write output to a different file |
| `--dry-run` | Show preview only, no JSON output |
| `--quiet` | Suppress preview, output only JSON |
| `--unwrap` | Reverse: remove `cmd.exe /c` wrapping |
| `--resolve` | Resolve commands to absolute paths via `PATH`/`PATHEXT` |
| `--no-color` | Disable colored output |

## What it does

- Wraps `.cmd` shim commands (`npx`, `uvx`, `pnpx`, `bunx`, `yarn`, `tsx`, etc.) with `cmd.exe /c`
- Skips native executables (`node`, `python`, `deno`, `bun`)
- Skips HTTP/SSE transport servers
- Skips already-wrapped commands (idempotent)
- Handles Zed's nested command format
- Preserves all extra fields (`env`, `disabled`, etc.)
- Supports Claude, VS Code, Cursor, Zed, Amazon Q, and Gemini configs

## License

MIT
