package transform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformServer(t *testing.T) {
	tests := []struct {
		name    string
		server  map[string]any
		resolve bool
		action  string
		changed bool
		wantCmd string
		wantArgs []any
	}{
		{
			name:    "wrap npx",
			server:  map[string]any{"command": "npx", "args": []any{"-y", "@modelcontextprotocol/server-filesystem"}},
			action:  "wrapped",
			changed: true,
			wantCmd: "cmd.exe",
			wantArgs: []any{"/c", "npx", "-y", "@modelcontextprotocol/server-filesystem"},
		},
		{
			name:    "wrap uvx",
			server:  map[string]any{"command": "uvx", "args": []any{"mcp-server-fetch"}},
			action:  "wrapped",
			changed: true,
			wantCmd: "cmd.exe",
			wantArgs: []any{"/c", "uvx", "mcp-server-fetch"},
		},
		{
			name:    "wrap bunx",
			server:  map[string]any{"command": "bunx", "args": []any{"some-server"}},
			action:  "wrapped",
			changed: true,
			wantCmd: "cmd.exe",
			wantArgs: []any{"/c", "bunx", "some-server"},
		},
		{
			name:    "wrap command with no args",
			server:  map[string]any{"command": "npx"},
			action:  "wrapped",
			changed: true,
			wantCmd: "cmd.exe",
			wantArgs: []any{"/c", "npx"},
		},
		{
			name:    "skip HTTP url",
			server:  map[string]any{"url": "http://localhost:3000/mcp"},
			action:  "skipped",
			changed: false,
		},
		{
			name:    "skip HTTP serverUrl",
			server:  map[string]any{"serverUrl": "http://localhost:3000/sse"},
			action:  "skipped",
			changed: false,
		},
		{
			name:    "skip SSE type",
			server:  map[string]any{"command": "npx", "type": "sse"},
			action:  "skipped",
			changed: false,
		},
		{
			name:    "skip already wrapped",
			server:  map[string]any{"command": "cmd.exe", "args": []any{"/c", "npx", "-y", "@pkg"}},
			action:  "already-wrapped",
			changed: false,
		},
		{
			name:    "skip CMD (uppercase)",
			server:  map[string]any{"command": "CMD", "args": []any{"/c", "npx"}},
			action:  "already-wrapped",
			changed: false,
		},
		{
			name:    "skip native node",
			server:  map[string]any{"command": "node", "args": []any{"server.js"}},
			action:  "native-exe",
			changed: false,
		},
		{
			name:    "skip native python",
			server:  map[string]any{"command": "python", "args": []any{"-m", "mcp_server"}},
			action:  "native-exe",
			changed: false,
		},
		{
			name:    "skip native deno",
			server:  map[string]any{"command": "deno", "args": []any{"run", "server.ts"}},
			action:  "native-exe",
			changed: false,
		},
		{
			name:    "skip no command",
			server:  map[string]any{"env": map[string]any{"KEY": "val"}},
			action:  "skipped",
			changed: false,
		},
		{
			name:    "preserve extra fields",
			server:  map[string]any{"command": "npx", "args": []any{"-y", "@pkg"}, "env": map[string]any{"KEY": "val"}, "disabled": false},
			action:  "wrapped",
			changed: true,
			wantCmd: "cmd.exe",
			wantArgs: []any{"/c", "npx", "-y", "@pkg"},
		},
		{
			name:    "case insensitive NPX",
			server:  map[string]any{"command": "NPX", "args": []any{"-y", "@pkg"}},
			action:  "wrapped",
			changed: true,
			wantCmd: "cmd.exe",
			wantArgs: []any{"/c", "NPX", "-y", "@pkg"},
		},
		{
			name:    "resolve mode with lookup",
			server:  map[string]any{"command": "npx", "args": []any{"-y", "@pkg"}},
			resolve: true,
			action:  "resolved",
			changed: true,
			wantCmd: "C:\\Program Files\\nodejs\\npx.cmd",
			wantArgs: []any{"-y", "@pkg"},
		},
		{
			name:    "resolve mode fallback to wrap",
			server:  map[string]any{"command": "npx", "args": []any{"-y", "@pkg"}},
			resolve: true,
			action:  "wrapped",
			changed: true,
			wantCmd: "cmd.exe",
			wantArgs: []any{"/c", "npx", "-y", "@pkg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lookupFn LookupFunc
			if tt.resolve && tt.action == "resolved" {
				lookupFn = func(cmd string) (string, error) {
					return `C:\Program Files\nodejs\npx.cmd`, nil
				}
			} else if tt.resolve {
				lookupFn = func(cmd string) (string, error) {
					return "", nil
				}
			}

			out, result := TransformServer("test", tt.server, tt.resolve, lookupFn)
			assert.Equal(t, tt.action, result.Action)
			assert.Equal(t, tt.changed, result.Changed)

			if tt.wantCmd != "" {
				assert.Equal(t, tt.wantCmd, out["command"])
			}
			if tt.wantArgs != nil {
				assert.Equal(t, tt.wantArgs, out["args"])
			}

			// Verify extra fields are preserved.
			if env, ok := tt.server["env"]; ok {
				assert.Equal(t, env, out["env"])
			}
		})
	}
}

func TestTransformServer_ZedNested(t *testing.T) {
	server := map[string]any{
		"command": map[string]any{
			"path": "npx",
			"args": []any{"-y", "@pkg"},
		},
	}

	out, result := TransformServer("zed-srv", server, false, nil)
	assert.Equal(t, "wrapped", result.Action)
	assert.True(t, result.Changed)

	cmdObj, ok := out["command"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "cmd.exe", cmdObj["path"])
	assert.Equal(t, []any{"/c", "npx", "-y", "@pkg"}, cmdObj["args"])
}

func TestUnwrapServer(t *testing.T) {
	tests := []struct {
		name     string
		server   map[string]any
		action   string
		changed  bool
		wantCmd  string
		wantArgs []any
		noArgs   bool
	}{
		{
			name:     "unwrap cmd.exe /c npx",
			server:   map[string]any{"command": "cmd.exe", "args": []any{"/c", "npx", "-y", "@pkg"}},
			action:   "unwrapped",
			changed:  true,
			wantCmd:  "npx",
			wantArgs: []any{"-y", "@pkg"},
		},
		{
			name:    "unwrap cmd.exe /c npx (no extra args)",
			server:  map[string]any{"command": "cmd.exe", "args": []any{"/c", "npx"}},
			action:  "unwrapped",
			changed: true,
			wantCmd: "npx",
			noArgs:  true,
		},
		{
			name:    "skip non-wrapped",
			server:  map[string]any{"command": "npx", "args": []any{"-y", "@pkg"}},
			action:  "skipped",
			changed: false,
		},
		{
			name:    "skip cmd.exe without /c",
			server:  map[string]any{"command": "cmd.exe", "args": []any{"npx"}},
			action:  "skipped",
			changed: false,
		},
		{
			name:    "skip cmd.exe with empty args",
			server:  map[string]any{"command": "cmd.exe", "args": []any{}},
			action:  "skipped",
			changed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, result := UnwrapServer("test", tt.server)
			assert.Equal(t, tt.action, result.Action)
			assert.Equal(t, tt.changed, result.Changed)

			if tt.wantCmd != "" {
				assert.Equal(t, tt.wantCmd, out["command"])
			}
			if tt.wantArgs != nil {
				assert.Equal(t, tt.wantArgs, out["args"])
			}
			if tt.noArgs {
				_, hasArgs := out["args"]
				assert.False(t, hasArgs, "args should be removed")
			}
		})
	}
}

func TestTransformAll(t *testing.T) {
	data := map[string]any{
		"mcpServers": map[string]any{
			"github": map[string]any{
				"command": "npx",
				"args":    []any{"-y", "@modelcontextprotocol/server-github"},
			},
			"filesystem": map[string]any{
				"command": "npx",
				"args":    []any{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
			},
			"weather": map[string]any{
				"url": "http://localhost:8080/mcp",
			},
		},
	}

	out, results := TransformAll(data, "mcpServers", false, false, nil)

	// Count changes.
	changed := 0
	skipped := 0
	for _, r := range results {
		if r.Changed {
			changed++
		} else {
			skipped++
		}
	}
	assert.Equal(t, 2, changed)
	assert.Equal(t, 1, skipped)

	// Verify output structure.
	servers := out["mcpServers"].(map[string]any)
	gh := servers["github"].(map[string]any)
	assert.Equal(t, "cmd.exe", gh["command"])
	assert.Equal(t, []any{"/c", "npx", "-y", "@modelcontextprotocol/server-github"}, gh["args"])

	weather := servers["weather"].(map[string]any)
	assert.Equal(t, "http://localhost:8080/mcp", weather["url"])
}
