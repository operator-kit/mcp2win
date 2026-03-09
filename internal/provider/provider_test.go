package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		args     []string
		wantName string
		wantN    int
	}{
		{[]string{"claude", "mcp", "add"}, "claude", 1},
		{[]string{"code", "--add-mcp"}, "vscode", 1},
		{[]string{"qchat", "mcp", "add"}, "qchat", 1},
		{[]string{"q", "mcp", "add"}, "q", 1},
		{[]string{"gemini", "mcp", "add"}, "gemini", 1},
		{[]string{"unknown"}, "", 0},
		{nil, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			p, n := DetectProvider(tt.args)
			assert.Equal(t, tt.wantN, n)
			if tt.wantName == "" {
				assert.Nil(t, p)
			} else {
				assert.Equal(t, tt.wantName, p.Name())
			}
		})
	}
}

func TestClaude_ParseArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantCmd string
		wantErr bool
	}{
		{
			name:    "basic",
			args:    []string{"mcp", "add", "github-server", "--", "npx", "-y", "@modelcontextprotocol/server-github"},
			wantCmd: "npx",
		},
		{
			name:    "with scope",
			args:    []string{"mcp", "add", "--scope", "user", "github-server", "--", "npx", "-y", "@pkg"},
			wantCmd: "npx",
		},
		{
			name:    "with env",
			args:    []string{"mcp", "add", "--env", "KEY=val", "srv", "--", "npx", "-y", "@pkg"},
			wantCmd: "npx",
		},
		{
			name:    "missing separator with unknown command",
			args:    []string{"mcp", "add", "srv", "custom-cmd"},
			wantErr: true,
		},
		{
			name:    "missing separator with known command (heuristic)",
			args:    []string{"mcp", "add", "srv", "npx", "-y", "@pkg"},
			wantCmd: "npx",
		},
	}

	c := &Claude{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := c.ParseArgs(tt.args)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantCmd, parsed.Command)
		})
	}
}

func TestClaude_FormatOutput(t *testing.T) {
	c := &Claude{}
	parsed := ParsedCLI{
		ServerName: "github-server",
		Command:    "npx",
		Args:       []string{"-y", "@pkg"},
	}
	transformed := map[string]any{
		"command": "cmd.exe",
		"args":    []any{"/c", "npx", "-y", "@pkg"},
	}
	output := c.FormatOutput(parsed, transformed)
	assert.Contains(t, output, "claude mcp add-json")
	assert.Contains(t, output, "github-server")
	assert.Contains(t, output, "cmd.exe")
}

func TestVSCode_ParseArgs(t *testing.T) {
	v := &VSCode{}
	args := []string{"--add-mcp", `{"name":"my-server","command":"npx","args":["-y","@pkg"]}`}
	parsed, err := v.ParseArgs(args)
	require.NoError(t, err)
	assert.Equal(t, "my-server", parsed.ServerName)
	assert.Equal(t, "npx", parsed.Command)
}

func TestVSCode_FormatOutput(t *testing.T) {
	v := &VSCode{}
	parsed := ParsedCLI{
		ServerName: "my-server",
		Extra:      map[string]any{"name": "my-server"},
	}
	transformed := map[string]any{
		"command": "cmd.exe",
		"args":    []any{"/c", "npx", "-y", "@pkg"},
	}
	output := v.FormatOutput(parsed, transformed)
	assert.Contains(t, output, "code --add-mcp")
	assert.Contains(t, output, "cmd.exe")
}

func TestAmazonQ_ParseArgs(t *testing.T) {
	a := &AmazonQ{cmd: "qchat"}
	args := []string{"mcp", "add", "--", "npx", "-y", "@pkg"}
	parsed, err := a.ParseArgs(args)
	require.NoError(t, err)
	assert.Equal(t, "npx", parsed.Command)
}

func TestAmazonQ_ParseArgs_NoSeparator_Heuristic(t *testing.T) {
	a := &AmazonQ{cmd: "qchat"}
	args := []string{"mcp", "add", "srv", "npx", "-y", "@pkg"}
	parsed, err := a.ParseArgs(args)
	require.NoError(t, err)
	assert.Equal(t, "srv", parsed.ServerName)
	assert.Equal(t, "npx", parsed.Command)
	assert.Equal(t, []string{"-y", "@pkg"}, parsed.Args)
}

func TestGemini_ParseArgs(t *testing.T) {
	g := &Gemini{}
	args := []string{"mcp", "add", "fetch-server", "--", "npx", "-y", "@pkg"}
	parsed, err := g.ParseArgs(args)
	require.NoError(t, err)
	assert.Equal(t, "fetch-server", parsed.ServerName)
	assert.Equal(t, "npx", parsed.Command)
}

func TestGemini_ParseArgs_NoSeparator_Heuristic(t *testing.T) {
	g := &Gemini{}
	args := []string{"mcp", "add", "fetch-server", "npx", "-y", "@pkg"}
	parsed, err := g.ParseArgs(args)
	require.NoError(t, err)
	assert.Equal(t, "fetch-server", parsed.ServerName)
	assert.Equal(t, "npx", parsed.Command)
	assert.Equal(t, []string{"-y", "@pkg"}, parsed.Args)
}

func TestClaude_ParseArgs_NoSeparator_WithScope(t *testing.T) {
	c := &Claude{}
	// Simulates: claude mcp add -s user playwright npx -y @playwright/mcp-server
	// (no '--', consumed by npm shim)
	args := []string{"mcp", "add", "-s", "user", "playwright", "npx", "-y", "@playwright/mcp-server"}
	parsed, err := c.ParseArgs(args)
	require.NoError(t, err)
	assert.Equal(t, "playwright", parsed.ServerName)
	assert.Equal(t, "user", parsed.Scope)
	assert.Equal(t, "npx", parsed.Command)
	assert.Equal(t, []string{"-y", "@playwright/mcp-server"}, parsed.Args)
}

func TestGemini_FormatOutput(t *testing.T) {
	g := &Gemini{}
	parsed := ParsedCLI{ServerName: "fetch-server"}
	transformed := map[string]any{
		"command": "cmd.exe",
		"args":    []any{"/c", "npx", "-y", "@pkg"},
	}
	output := g.FormatOutput(parsed, transformed)
	assert.Contains(t, output, "gemini mcp add fetch-server -- cmd.exe /c npx -y @pkg")
	assert.Contains(t, output, "natively")
}

// --- ExecArgs tests ---

func TestClaude_ExecArgs(t *testing.T) {
	c := &Claude{}
	parsed := ParsedCLI{
		ServerName: "github-server",
		Command:    "npx",
		Args:       []string{"-y", "@pkg"},
	}
	transformed := map[string]any{
		"command": "cmd.exe",
		"args":    []any{"/c", "npx", "-y", "@pkg"},
	}

	exe, args := c.ExecArgs(parsed, transformed)
	assert.Equal(t, "claude", exe)
	assert.Equal(t, "mcp", args[0])
	assert.Equal(t, "add-json", args[1])
	assert.Equal(t, "github-server", args[2])
	// JSON arg: no surrounding single quotes
	jsonArg := args[3]
	assert.NotContains(t, jsonArg, "'")
	assert.Contains(t, jsonArg, "cmd.exe")
	assert.Contains(t, jsonArg, "/c")
}

func TestClaude_ExecArgs_WithScope(t *testing.T) {
	c := &Claude{}
	parsed := ParsedCLI{
		ServerName: "srv",
		Scope:      "user",
	}
	transformed := map[string]any{
		"command": "cmd.exe",
		"args":    []any{"/c", "npx"},
	}

	exe, args := c.ExecArgs(parsed, transformed)
	assert.Equal(t, "claude", exe)
	assert.Contains(t, args, "--scope")
	assert.Contains(t, args, "user")
}

func TestClaude_ExecArgs_WithEnv(t *testing.T) {
	c := &Claude{}
	parsed := ParsedCLI{
		ServerName: "srv",
		EnvVars:    map[string]string{"TOKEN": "abc"},
	}
	transformed := map[string]any{
		"command": "cmd.exe",
		"args":    []any{"/c", "npx"},
	}

	_, args := c.ExecArgs(parsed, transformed)
	jsonArg := args[len(args)-1]
	assert.Contains(t, jsonArg, "TOKEN")
	assert.Contains(t, jsonArg, "abc")
}

func TestVSCode_ExecArgs(t *testing.T) {
	v := &VSCode{}
	parsed := ParsedCLI{
		ServerName: "my-server",
		Extra:      map[string]any{"name": "my-server"},
	}
	transformed := map[string]any{
		"command": "cmd.exe",
		"args":    []any{"/c", "npx", "-y", "@pkg"},
	}

	exe, args := v.ExecArgs(parsed, transformed)
	assert.Equal(t, "code", exe)
	assert.Equal(t, "--add-mcp", args[0])
	// No surrounding quotes on JSON
	assert.NotContains(t, args[1], "'")
	assert.Contains(t, args[1], "cmd.exe")
}

func TestAmazonQ_ExecArgs(t *testing.T) {
	a := &AmazonQ{cmd: "qchat"}
	parsed := ParsedCLI{ServerName: "srv"}
	transformed := map[string]any{
		"command": "cmd.exe",
		"args":    []any{"/c", "npx", "-y", "@pkg"},
	}

	exe, args := a.ExecArgs(parsed, transformed)
	assert.Equal(t, "qchat", exe)
	assert.Equal(t, []string{"mcp", "add", "srv", "--", "cmd.exe", "/c", "npx", "-y", "@pkg"}, args)
}

func TestGemini_ExecArgs(t *testing.T) {
	g := &Gemini{}
	parsed := ParsedCLI{ServerName: "fetch-server"}
	transformed := map[string]any{
		"command": "cmd.exe",
		"args":    []any{"/c", "npx", "-y", "@pkg"},
	}

	exe, args := g.ExecArgs(parsed, transformed)
	assert.Equal(t, "gemini", exe)
	assert.Equal(t, []string{"mcp", "add", "fetch-server", "--", "cmd.exe", "/c", "npx", "-y", "@pkg"}, args)
}
