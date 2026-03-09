package cmd

import (
	"bytes"
	"testing"

	"github.com/operator-kit/mcp2win/internal/config"
	"github.com/stretchr/testify/assert"
)

// --- Dry-run tests (print command, don't execute) ---

func TestRunCLI_Claude_DryRun(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"claude", "mcp", "add", "github-server", "--", "npx", "-y", "@modelcontextprotocol/server-github"}

	code := runCLI(args, Flags{DryRun: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "claude mcp add-json")
	assert.Contains(t, stdout.String(), "github-server")
	assert.Contains(t, stdout.String(), "cmd.exe")
}

func TestRunCLI_ClaudeWithScope_DryRun(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"claude", "mcp", "add", "--scope", "user", "srv", "--", "npx", "-y", "@pkg"}

	code := runCLI(args, Flags{DryRun: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "--scope user")
}

func TestRunCLI_VSCode_DryRun(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"code", "--add-mcp", `{"name":"my-server","command":"npx","args":["-y","@pkg"]}`}

	code := runCLI(args, Flags{DryRun: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "code --add-mcp")
	assert.Contains(t, stdout.String(), "cmd.exe")
}

func TestRunCLI_AmazonQ_DryRun(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"qchat", "mcp", "add", "--", "npx", "-y", "@pkg"}

	code := runCLI(args, Flags{DryRun: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "qchat mcp add -- cmd.exe /c npx -y @pkg")
}

func TestRunCLI_Gemini_DryRun(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"gemini", "mcp", "add", "srv", "--", "npx", "-y", "@pkg"}

	code := runCLI(args, Flags{DryRun: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "gemini mcp add srv -- cmd.exe /c npx -y @pkg")
}

func TestRunCLI_Generic_DryRun(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"somecli", "mcp", "add", "--", "npx", "-y", "@pkg"}

	code := runCLI(args, Flags{DryRun: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "cmd.exe /c npx")
}

// --- Skip cases (no wrapping needed, no exec) ---

func TestRunCLI_NativeExe(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"claude", "mcp", "add", "srv", "--", "node", "server.js"}

	code := runCLI(args, Flags{}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr.String(), "native executable")
}

func TestRunCLI_AlreadyWrapped(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"claude", "mcp", "add", "srv", "--", "cmd.exe", "/c", "npx"}

	code := runCLI(args, Flags{}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr.String(), "already wrapped")
}

func TestRunCLI_GenericNoSeparator(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	args := []string{"somecli", "npx", "-y"}

	code := runCLI(args, Flags{}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "unknown provider")
}

// --- Exec tests (default behavior with --yes to skip prompt) ---

func withMockExec(t *testing.T, fn func(capturedExe *string, capturedArgs *[]string)) {
	t.Helper()
	orig := ExecFn
	t.Cleanup(func() { ExecFn = orig })

	var capturedExe string
	var capturedArgs []string
	ExecFn = func(exe string, args []string) int {
		capturedExe = exe
		capturedArgs = args
		return 0
	}
	fn(&capturedExe, &capturedArgs)
}

func TestRunCLI_Claude_Exec(t *testing.T) {
	withMockExec(t, func(exe *string, args *[]string) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		cliArgs := []string{"claude", "mcp", "add", "github-server", "--", "npx", "-y", "@modelcontextprotocol/server-github"}

		code := runCLI(cliArgs, Flags{Yes: true}, defaultCfg(), stdout, stderr)
		assert.Equal(t, 0, code)
		assert.Equal(t, "claude", *exe)
		assert.Contains(t, *args, "mcp")
		assert.Contains(t, *args, "add-json")
		assert.Contains(t, *args, "github-server")
	})
}

func TestRunCLI_ClaudeWithScope_Exec(t *testing.T) {
	withMockExec(t, func(exe *string, args *[]string) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		cliArgs := []string{"claude", "mcp", "add", "--scope", "user", "srv", "--", "npx", "-y", "@pkg"}

		code := runCLI(cliArgs, Flags{Yes: true}, defaultCfg(), stdout, stderr)
		assert.Equal(t, 0, code)
		assert.Equal(t, "claude", *exe)
		assert.Contains(t, *args, "--scope")
		assert.Contains(t, *args, "user")
	})
}

func TestRunCLI_VSCode_Exec(t *testing.T) {
	withMockExec(t, func(exe *string, args *[]string) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		cliArgs := []string{"code", "--add-mcp", `{"name":"my-server","command":"npx","args":["-y","@pkg"]}`}

		code := runCLI(cliArgs, Flags{Yes: true}, defaultCfg(), stdout, stderr)
		assert.Equal(t, 0, code)
		assert.Equal(t, "code", *exe)
		assert.Equal(t, "--add-mcp", (*args)[0])
		// JSON arg should NOT have surrounding single quotes.
		assert.NotContains(t, (*args)[1], "'")
		assert.Contains(t, (*args)[1], "cmd.exe")
	})
}

func TestRunCLI_AmazonQ_Exec(t *testing.T) {
	withMockExec(t, func(exe *string, args *[]string) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		cliArgs := []string{"qchat", "mcp", "add", "--", "npx", "-y", "@pkg"}

		code := runCLI(cliArgs, Flags{Yes: true}, defaultCfg(), stdout, stderr)
		assert.Equal(t, 0, code)
		assert.Equal(t, "qchat", *exe)
		assert.Contains(t, *args, "cmd.exe")
		assert.Contains(t, *args, "/c")
		assert.Contains(t, *args, "npx")
	})
}

func TestRunCLI_Gemini_Exec(t *testing.T) {
	withMockExec(t, func(exe *string, args *[]string) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		cliArgs := []string{"gemini", "mcp", "add", "srv", "--", "npx", "-y", "@pkg"}

		code := runCLI(cliArgs, Flags{Yes: true}, defaultCfg(), stdout, stderr)
		assert.Equal(t, 0, code)
		assert.Equal(t, "gemini", *exe)
		assert.Contains(t, *args, "srv")
		assert.Contains(t, *args, "cmd.exe")
	})
}

func TestRunCLI_Generic_Exec(t *testing.T) {
	withMockExec(t, func(exe *string, args *[]string) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		cliArgs := []string{"somecli", "mcp", "add", "--", "npx", "-y", "@pkg"}

		code := runCLI(cliArgs, Flags{Yes: true}, defaultCfg(), stdout, stderr)
		assert.Equal(t, 0, code)
		assert.Equal(t, "somecli", *exe)
		assert.Contains(t, *args, "cmd.exe")
		assert.Contains(t, *args, "/c")
	})
}

func TestRunCLI_Quiet_NoPreview(t *testing.T) {
	withMockExec(t, func(exe *string, args *[]string) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		cliArgs := []string{"claude", "mcp", "add", "srv", "--", "npx", "-y", "@pkg"}

		code := runCLI(cliArgs, Flags{Quiet: true, Yes: true}, defaultCfg(), stdout, stderr)
		assert.Equal(t, 0, code)
		assert.Equal(t, "claude", *exe)
		assert.Empty(t, stderr.String())
	})
}

func TestRunCLI_ExitCodePropagation(t *testing.T) {
	orig := ExecFn
	t.Cleanup(func() { ExecFn = orig })

	ExecFn = func(exe string, args []string) int {
		return 42
	}

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	cliArgs := []string{"claude", "mcp", "add", "srv", "--", "npx", "-y", "@pkg"}

	code := runCLI(cliArgs, Flags{Yes: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 42, code)
}

func TestRunCLI_ConfigAlwaysExec(t *testing.T) {
	withMockExec(t, func(exe *string, args *[]string) {
		stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
		cliArgs := []string{"claude", "mcp", "add", "srv", "--", "npx", "-y", "@pkg"}

		// Config preference set, no --yes flag needed.
		cfg := &config.Config{AlwaysExecCLI: true}
		code := runCLI(cliArgs, Flags{}, cfg, stdout, stderr)
		assert.Equal(t, 0, code)
		assert.Equal(t, "claude", *exe)
	})
}
