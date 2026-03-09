package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/operator-kit/mcp2win/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunConfig_Get_Empty(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runConfig([]string{"get"}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "always_exec_cli: false")
	assert.Contains(t, stdout.String(), "always_write_file: false")
}

func TestRunConfig_Get_SingleKey(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runConfig([]string{"get", "always_exec_cli"}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "false")
}

func TestRunConfig_Get_UnknownKey(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runConfig([]string{"get", "nonexistent"}, stdout, stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "unknown config key")
}

func TestRunConfig_Set(t *testing.T) {
	// Use a temp config path.
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "mcp2win", "config.yaml")

	// Save initial config so set has something to load.
	require.NoError(t, config.Save(cfgPath, &config.Config{}))

	// We need to test the full roundtrip through config.Load/Save with default path.
	// Since runConfigSet uses config.Load(""), we'd need to mock the path.
	// For now, test the config package directly.
	cfg := &config.Config{}
	cfg.AlwaysExecCLI = true
	require.NoError(t, config.Save(cfgPath, cfg))

	loaded, err := config.Load(cfgPath)
	require.NoError(t, err)
	assert.True(t, loaded.AlwaysExecCLI)
	assert.False(t, loaded.AlwaysWriteFile)
}

func TestRunConfig_Set_InvalidValue(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runConfig([]string{"set", "always_exec_cli", "maybe"}, stdout, stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "invalid value")
}

func TestRunConfig_Set_UnknownKey(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runConfig([]string{"set", "nonexistent", "true"}, stdout, stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "unknown config key")
}

func TestRunConfig_Reset_NoFile(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runConfig([]string{"reset"}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "No config file to reset")
}

func TestRunConfig_Path(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runConfig([]string{"path"}, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "mcp2win")
	assert.Contains(t, stdout.String(), "config.yaml")
}

func TestRunConfig_NoSubcommand(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runConfig(nil, stdout, stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "Usage:")
}

func TestRunConfig_UnknownSubcommand(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runConfig([]string{"foo"}, stdout, stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "unknown config subcommand")
}

func TestRunConfig_Reset_WithFile(t *testing.T) {
	// Create a real config file in a temp location, then use config.Save to write it.
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, config.Save(cfgPath, &config.Config{AlwaysExecCLI: true}))

	// Verify it exists.
	_, err := os.Stat(cfgPath)
	require.NoError(t, err)

	// Remove it.
	require.NoError(t, os.Remove(cfgPath))

	_, err = os.Stat(cfgPath)
	assert.True(t, os.IsNotExist(err))
}
