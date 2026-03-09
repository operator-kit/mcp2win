package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_NoFile(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
	require.NoError(t, err)
	assert.False(t, cfg.AlwaysExecCLI)
	assert.False(t, cfg.AlwaysWriteFile)
}

func TestSaveAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mcp2win", "config.yaml")
	cfg := &Config{AlwaysExecCLI: true, AlwaysWriteFile: false}

	require.NoError(t, Save(path, cfg))

	loaded, err := Load(path)
	require.NoError(t, err)
	assert.True(t, loaded.AlwaysExecCLI)
	assert.False(t, loaded.AlwaysWriteFile)
}

func TestSaveAndLoad_AllTrue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg := &Config{AlwaysExecCLI: true, AlwaysWriteFile: true}

	require.NoError(t, Save(path, cfg))

	loaded, err := Load(path)
	require.NoError(t, err)
	assert.True(t, loaded.AlwaysExecCLI)
	assert.True(t, loaded.AlwaysWriteFile)
}

func TestSave_CreatesParentDirs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "deep", "nested", "config.yaml")
	require.NoError(t, Save(path, &Config{}))

	_, err := os.Stat(path)
	assert.NoError(t, err)
}

func TestDefaultPath(t *testing.T) {
	p := DefaultPath()
	assert.Contains(t, p, "mcp2win")
	assert.Contains(t, p, "config.yaml")
}
