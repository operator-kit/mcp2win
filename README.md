# mcp2win

Finally, MCP servers that just work on Windows.

Every MCP server README assumes macOS or Linux. You copy the config, paste it in, and nothing happens — because `npx`, `uvx`, and friends are `.cmd` batch shims on Windows, not real executables. `mcp2win` fixes this by wrapping commands with `cmd.exe /c` so your MCP servers actually start.

Works with **Claude Code/Desktop**, **VS Code**, **Cursor**, **Zed**, **Amazon Q**, and **Gemini CLI**

Copy any install command, prefix with `mcp2win`, confirm — done. Or convert online at [mcp2win.sh](https://mcp2win.sh).

## Quick start

No install needed — just run with `npx`:

```bash
# Copy an MCP server's install command, prefix with mcp2win — confirm and it runs
npx @operatorkit/mcp2win claude mcp add github-server -- npx -y @modelcontextprotocol/server-github
# Shows the converted command, asks to confirm, then runs it

# Fix a config file (confirms before writing, creates .bak backup)
npx @operatorkit/mcp2win claude_desktop_config.json

# Convert inline JSON (outputs to stdout, no confirmation needed)
npx @operatorkit/mcp2win '{"command":"npx","args":["-y","@pkg"]}'
```

## Install

If you use it often, install globally:

```bash
npm i -g @operatorkit/mcp2win

# or with go
go install github.com/operator-kit/mcp2win/cmd/mcp2win@latest

# or download a binary from https://github.com/operator-kit/mcp2win/releases
```

## Usage

### CLI command translation

Paste any provider's `mcp add` command — `mcp2win` converts it and **runs it after confirmation**:

```bash
# Claude — converts to add-json, confirms, runs
mcp2win claude mcp add github-server -- npx -y @modelcontextprotocol/server-github
# Execute? [y]es / [n]o / [a]lways: y

# VS Code
mcp2win code --add-mcp '{"name":"my-server","command":"npx","args":["-y","@pkg"]}'

# Amazon Q
mcp2win qchat mcp add -- npx -y @pkg

# Gemini
mcp2win gemini mcp add fetch-server -- npx -y @pkg

# Skip confirmation (non-interactive / CI)
mcp2win -y claude mcp add github-server -- npx -y @modelcontextprotocol/server-github

# Preview only, don't run
mcp2win --dry-run claude mcp add github-server -- npx -y @modelcontextprotocol/server-github
```

Answering **always** saves your preference — no more prompts for future CLI commands.

### File transformation

```bash
# Preview + confirm write (creates .bak backup)
mcp2win claude_desktop_config.json
# Write changes to claude_desktop_config.json? [y]es / [n]o / [a]lways: y

# Skip confirmation
mcp2win -y claude_desktop_config.json

# Write without backup
mcp2win -y --no-backup claude_desktop_config.json

# Write to a different file
mcp2win -y -o windows_config.json claude_desktop_config.json

# Preview only
mcp2win --dry-run claude_desktop_config.json
```

### JSON conversion

Transform inline JSON or pipe from stdin (outputs to stdout, no confirmation):

```bash
# Inline
mcp2win '{"command":"npx","args":["-y","@modelcontextprotocol/server-github"]}'

# Stdin
cat claude_desktop_config.json | mcp2win

# Full config
mcp2win '{"mcpServers":{"s1":{"command":"npx","args":["-y","@pkg"]}}}'
```

### Updating

mcp2win checks for updates in the background and notifies you when a new version is available. To update:

```bash
mcp2win update
```

Disable background checks with `MCP2WIN_NO_UPDATE_CHECK=1`.

### Preferences

When you answer **always** at a prompt, the preference is saved. You can also manage preferences directly:

```bash
mcp2win config get                              # show all preferences
mcp2win config set always_exec_cli true         # skip CLI confirmation
mcp2win config set always_write_file true       # skip file confirmation
mcp2win config reset                            # reset all preferences
mcp2win config path                             # show config file location
```

## Flags

| Flag | Description |
|---|---|
| `--yes`, `-y` | Skip confirmation prompt (non-interactive mode) |
| `--dry-run` | Preview only, no action |
| `--quiet` | Suppress preview output |
| `--no-backup` | Skip `.bak` backup when writing files |
| `-o <path>` | Write output to a different file |
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
