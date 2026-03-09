package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/operator-kit/mcp2win/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultCfg() *config.Config {
	return &config.Config{}
}

func TestRunFile_Claude_DryRun(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{"../../testdata/claude.json"}, Flags{DryRun: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Empty(t, stdout.String())
	assert.Contains(t, stderr.String(), "Changes:")
}

func TestRunFile_Claude_OutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output.json")

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{"../../testdata/claude.json"}, Flags{Output: outFile, Yes: true, Quiet: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)

	content, err := os.ReadFile(outFile)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(content, &out))
	servers := out["mcpServers"].(map[string]any)

	gh := servers["github-server"].(map[string]any)
	assert.Equal(t, "cmd.exe", gh["command"])

	weather := servers["weather-api"].(map[string]any)
	assert.Equal(t, "http://localhost:8080/mcp", weather["url"])
}

func TestRunFile_Write(t *testing.T) {
	tmpDir := t.TempDir()
	src, err := os.ReadFile("../../testdata/claude.json")
	require.NoError(t, err)

	tmpFile := filepath.Join(tmpDir, "test.json")
	require.NoError(t, os.WriteFile(tmpFile, src, 0644))

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{tmpFile}, Flags{Yes: true, Quiet: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)

	// Check .bak was created.
	bakFile := tmpFile + ".bak"
	bakContent, err := os.ReadFile(bakFile)
	require.NoError(t, err)
	assert.Equal(t, src, bakContent)

	// Check file was modified.
	modified, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(modified, &out))
	servers := out["mcpServers"].(map[string]any)
	gh := servers["github-server"].(map[string]any)
	assert.Equal(t, "cmd.exe", gh["command"])
}

func TestRunFile_WriteNoBackup(t *testing.T) {
	tmpDir := t.TempDir()
	src, err := os.ReadFile("../../testdata/single_server.json")
	require.NoError(t, err)

	tmpFile := filepath.Join(tmpDir, "test.json")
	require.NoError(t, os.WriteFile(tmpFile, src, 0644))

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{tmpFile}, Flags{Yes: true, NoBackup: true, Quiet: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)

	// No .bak should exist.
	_, err = os.Stat(tmpFile + ".bak")
	assert.True(t, os.IsNotExist(err))
}

func TestRunFile_DryRun(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{"../../testdata/claude.json"}, Flags{DryRun: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Empty(t, stdout.String()) // No JSON output.
	assert.Contains(t, stderr.String(), "Changes:")
}

func TestRunFile_MixedTransports(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output.json")

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{"../../testdata/mixed_transports.json"}, Flags{Output: outFile, Yes: true, Quiet: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)

	content, err := os.ReadFile(outFile)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(content, &out))
	servers := out["mcpServers"].(map[string]any)

	// stdio-server should be wrapped.
	stdio := servers["stdio-server"].(map[string]any)
	assert.Equal(t, "cmd.exe", stdio["command"])

	// http-server should be untouched.
	http := servers["http-server"].(map[string]any)
	assert.Equal(t, "http://localhost:3000/mcp", http["url"])

	// native-server should be untouched.
	native := servers["native-server"].(map[string]any)
	assert.Equal(t, "node", native["command"])
}

func TestRunFile_AlreadyWrapped(t *testing.T) {
	// All servers already wrapped — no changes needed.
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{"../../testdata/already_wrapped.json"}, Flags{Yes: true, Quiet: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	// No output since nothing changed.
	assert.Empty(t, stdout.String())
}

func TestRunFile_WriteIncrementsBackup(t *testing.T) {
	tmpDir := t.TempDir()
	src, err := os.ReadFile("../../testdata/claude.json")
	require.NoError(t, err)

	tmpFile := filepath.Join(tmpDir, "test.json")
	require.NoError(t, os.WriteFile(tmpFile, src, 0644))

	// Pre-create .bak.
	existingBak := []byte("existing backup content")
	require.NoError(t, os.WriteFile(tmpFile+".bak", existingBak, 0644))

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{tmpFile}, Flags{Yes: true, Quiet: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)

	// Original .bak should be untouched.
	bak1, err := os.ReadFile(tmpFile + ".bak")
	require.NoError(t, err)
	assert.Equal(t, existingBak, bak1)

	// New backup should be .bak2.
	bak2, err := os.ReadFile(tmpFile + ".bak2")
	require.NoError(t, err)
	assert.Equal(t, src, bak2)
}

func TestRunFile_WriteIncrementsMultiple(t *testing.T) {
	tmpDir := t.TempDir()
	src, err := os.ReadFile("../../testdata/single_server.json")
	require.NoError(t, err)

	tmpFile := filepath.Join(tmpDir, "test.json")
	require.NoError(t, os.WriteFile(tmpFile, src, 0644))

	// Pre-create .bak and .bak2.
	require.NoError(t, os.WriteFile(tmpFile+".bak", []byte("bak1"), 0644))
	require.NoError(t, os.WriteFile(tmpFile+".bak2", []byte("bak2"), 0644))

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{tmpFile}, Flags{Yes: true, Quiet: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)

	// .bak and .bak2 untouched.
	b1, _ := os.ReadFile(tmpFile + ".bak")
	assert.Equal(t, []byte("bak1"), b1)
	b2, _ := os.ReadFile(tmpFile + ".bak2")
	assert.Equal(t, []byte("bak2"), b2)

	// New backup at .bak3.
	b3, err := os.ReadFile(tmpFile + ".bak3")
	require.NoError(t, err)
	assert.Equal(t, src, b3)
}

func TestRunFile_WriteShowsBackupPath(t *testing.T) {
	tmpDir := t.TempDir()
	src, err := os.ReadFile("../../testdata/single_server.json")
	require.NoError(t, err)

	tmpFile := filepath.Join(tmpDir, "test.json")
	require.NoError(t, os.WriteFile(tmpFile, src, 0644))
	require.NoError(t, os.WriteFile(tmpFile+".bak", []byte("old"), 0644))

	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{tmpFile}, Flags{Yes: true}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr.String(), ".bak2")
}

func TestRunFile_ConfigAlwaysWrite(t *testing.T) {
	tmpDir := t.TempDir()
	src, err := os.ReadFile("../../testdata/single_server.json")
	require.NoError(t, err)

	tmpFile := filepath.Join(tmpDir, "test.json")
	require.NoError(t, os.WriteFile(tmpFile, src, 0644))

	// Config has always_write_file = true, so no prompt needed.
	cfg := &config.Config{AlwaysWriteFile: true}
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{tmpFile}, Flags{Quiet: true}, cfg, stdout, stderr)
	assert.Equal(t, 0, code)

	// File should be modified.
	modified, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	assert.Contains(t, string(modified), "cmd.exe")
}

func TestNextBackupPath(t *testing.T) {
	tmpDir := t.TempDir()
	base := filepath.Join(tmpDir, "config.json")

	// No existing backups → .bak
	assert.Equal(t, base+".bak", nextBackupPath(base))

	// Create .bak → next should be .bak2
	os.WriteFile(base+".bak", []byte("x"), 0644)
	assert.Equal(t, base+".bak2", nextBackupPath(base))

	// Create .bak2 → next should be .bak3
	os.WriteFile(base+".bak2", []byte("x"), 0644)
	assert.Equal(t, base+".bak3", nextBackupPath(base))

	// Create .bak3 → next should be .bak4
	os.WriteFile(base+".bak3", []byte("x"), 0644)
	assert.Equal(t, base+".bak4", nextBackupPath(base))
}

func TestRunFile_NotFound(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runFile([]string{"nonexistent.json"}, Flags{}, defaultCfg(), stdout, stderr)
	assert.Equal(t, 1, code)
}
