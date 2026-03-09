# mcp2win

Finally, MCP servers that just work on Windows.

Every MCP server README assumes macOS or Linux. You copy the config, paste it in, and nothing happens — because `npx`, `uvx`, and friends are `.cmd` batch shims on Windows, not real executables. `mcp2win` fixes this by wrapping commands with `cmd.exe /c` so your MCP servers actually start.

Works with **Claude Code/Desktop**, **VS Code**, **Cursor**, **Zed**, **Amazon Q**, and **Gemini CLI**

Alternatively convert online at [mcp2win.sh](https://mcp2win.sh).

## Quick start

No install needed — just prefix with `npx`. Copy any MCP server's install command, add `npx @operatorkit/mcp2win` in front, confirm, done:

```bash
# Converts and runs after confirmation
npx @operatorkit/mcp2win claude mcp add github-server -- npx -y @modelcontextprotocol/server-github
# Execute? [y]es / [n]o / [a]lways: y

# Fix a config file (confirms before writing)
npx @operatorkit/mcp2win claude_desktop_config.json

npx @operatorkit/mcp2win '{"command":"npx","args":["-y","@pkg"]}'
```

## Install globally

```bash
npm i -g @operatorkit/mcp2win
```

Then use directly:

```bash
mcp2win claude mcp add github-server -- npx -y @modelcontextprotocol/server-github
mcp2win claude_desktop_config.json
mcp2win '{"command":"npx","args":["-y","@pkg"]}'
```

## Usage

**Translate & run CLI commands (confirms before executing):**

```bash
mcp2win claude mcp add srv -- npx -y @pkg
mcp2win code --add-mcp '{"name":"srv","command":"npx","args":["-y","@pkg"]}'
mcp2win qchat mcp add -- npx -y @pkg
mcp2win gemini mcp add srv -- npx -y @pkg
mcp2win -y claude mcp add srv -- npx -y @pkg      # skip confirmation
mcp2win --dry-run claude mcp add srv -- npx -y @pkg  # preview only
```

**Transform config files (confirms before writing):**

```bash
mcp2win config.json                        # confirm + write (creates .bak backup)
mcp2win -y config.json                     # skip confirmation
mcp2win -o windows_config.json config.json # write to different file
mcp2win --dry-run config.json              # preview only
```

**Convert JSON (inline or stdin):**

```bash
mcp2win '{"command":"npx","args":["-y","@pkg"]}'
cat config.json | mcp2win
```

**Update:**

```bash
mcp2win update                                  # update to latest version
```

**Preferences:**

```bash
mcp2win config get                              # show preferences
mcp2win config set always_exec_cli true         # skip CLI confirmation
mcp2win config set always_write_file true       # skip file confirmation
mcp2win config reset                            # reset all
```

## Flags

| Flag | Description |
|---|---|
| `--yes`, `-y` | Skip confirmation prompt |
| `--dry-run` | Preview only, no action |
| `--quiet` | Suppress preview output |
| `--no-backup` | Skip `.bak` when writing files |
| `-o <path>` | Write to a different file |
| `--unwrap` | Reverse: remove `cmd.exe /c` wrapping |
| `--resolve` | Resolve commands to absolute paths via PATH/PATHEXT |
| `--no-color` | Disable colored output |

## What it does

- Wraps shim commands (`npx`, `uvx`, `pnpx`, `bunx`, `yarn`, `tsx`, etc.) with `cmd.exe /c`
- Skips native executables (`node`, `python`, `deno`, `bun`)
- Skips HTTP/SSE transports and already-wrapped commands
- Preserves all extra fields (`env`, `disabled`, etc.)
- Idempotent — safe to run multiple times

[Full documentation on GitHub](https://github.com/operator-kit/mcp2win)

## License

MIT
