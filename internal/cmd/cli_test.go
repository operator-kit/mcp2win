package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunCLI_Claude(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"claude", "mcp", "add", "github-server", "--", "npx", "-y", "@modelcontextprotocol/server-github"}

	code := runCLI(args, Flags{}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "claude mcp add-json")
	assert.Contains(t, stdout.String(), "github-server")
	assert.Contains(t, stdout.String(), "cmd.exe")
}

func TestRunCLI_ClaudeWithScope(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"claude", "mcp", "add", "--scope", "user", "srv", "--", "npx", "-y", "@pkg"}

	code := runCLI(args, Flags{}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "--scope user")
}

func TestRunCLI_VSCode(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"code", "--add-mcp", `{"name":"my-server","command":"npx","args":["-y","@pkg"]}`}

	code := runCLI(args, Flags{}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "code --add-mcp")
	assert.Contains(t, stdout.String(), "cmd.exe")
}

func TestRunCLI_AmazonQ(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"qchat", "mcp", "add", "--", "npx", "-y", "@pkg"}

	code := runCLI(args, Flags{}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "qchat mcp add -- cmd.exe /c npx -y @pkg")
}

func TestRunCLI_Gemini(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"gemini", "mcp", "add", "srv", "--", "npx", "-y", "@pkg"}

	code := runCLI(args, Flags{}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "gemini mcp add srv -- cmd.exe /c npx -y @pkg")
}

func TestRunCLI_NativeExe(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"claude", "mcp", "add", "srv", "--", "node", "server.js"}

	code := runCLI(args, Flags{}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr.String(), "native executable")
}

func TestRunCLI_AlreadyWrapped(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"claude", "mcp", "add", "srv", "--", "cmd.exe", "/c", "npx"}

	code := runCLI(args, Flags{}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr.String(), "already wrapped")
}

func TestRunCLI_GenericProvider(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"somecli", "mcp", "add", "--", "npx", "-y", "@pkg"}

	code := runCLI(args, Flags{}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "cmd.exe /c npx")
}

func TestRunCLI_GenericNoSeparator(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"somecli", "npx", "-y"}

	code := runCLI(args, Flags{}, stdout, stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "unknown provider")
}
