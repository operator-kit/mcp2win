# mcp2win

Finally, MCP servers that just work on Windows.

Every MCP server README assumes macOS or Linux. You copy the config, paste it in, and nothing happens — because `npx`, `uvx`, and friends are `.cmd` batch shims on Windows, not real executables. `mcp2win` fixes this by wrapping commands with `cmd.exe /c` so your MCP servers actually start.

Works with **Claude Code/Desktop**, **VS Code**, **Cursor**, **Zed**, **Amazon Q**, and **Gemini CLI**

## Quick start

No install needed — just prefix with `npx`:

```bash
npx @operatorkit/mcp2win --write claude_desktop_config.json
npx @operatorkit/mcp2win claude mcp add github-server -- npx -y @modelcontextprotocol/server-github
npx @operatorkit/mcp2win '{"command":"npx","args":["-y","@pkg"]}'
```

## Install globally

```bash
npm i -g @operatorkit/mcp2win
```

Then use directly:

```bash
mcp2win --write claude_desktop_config.json
mcp2win claude mcp add github-server -- npx -y @modelcontextprotocol/server-github
mcp2win '{"command":"npx","args":["-y","@pkg"]}'
```

## Usage

**Transform config files:**

```bash
mcp2win config.json                        # preview changes
mcp2win --write config.json                # apply (creates .bak backup)
mcp2win -o windows_config.json config.json # write to different file
```

**Translate CLI commands:**

```bash
mcp2win claude mcp add srv -- npx -y @pkg
mcp2win code --add-mcp '{"name":"srv","command":"npx","args":["-y","@pkg"]}'
mcp2win qchat mcp add -- npx -y @pkg
mcp2win gemini mcp add srv -- npx -y @pkg
```

**Convert JSON (inline or stdin):**

```bash
mcp2win '{"command":"npx","args":["-y","@pkg"]}'
cat config.json | mcp2win
```

## Flags

| Flag | Description |
|---|---|
| `--write` | Write changes back to file (creates `.bak` backup) |
| `--no-backup` | Skip `.bak` when using `--write` |
| `-o <path>` | Write to a different file |
| `--dry-run` | Preview only, no output |
| `--quiet` | JSON output only, no preview |
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
