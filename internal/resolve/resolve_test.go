package resolve

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCommand(t *testing.T) {
	tmpDir := t.TempDir()

	// Create fake executables.
	for _, name := range []string{"npx.cmd", "uvx.exe", "node.exe"} {
		f, err := os.Create(filepath.Join(tmpDir, name))
		require.NoError(t, err)
		f.Close()
	}

	pathDirs := []string{tmpDir}
	pathExt := []string{".cmd", ".exe", ".bat"}

	tests := []struct {
		name    string
		cmd     string
		wantErr bool
		wantAbs bool
	}{
		{
			name:    "resolve npx to npx.cmd",
			cmd:     "npx",
			wantAbs: true,
		},
		{
			name:    "resolve uvx to uvx.exe",
			cmd:     "uvx",
			wantAbs: true,
		},
		{
			name:    "resolve node to node.exe",
			cmd:     "node",
			wantAbs: true,
		},
		{
			name:    "not found",
			cmd:     "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ResolveCommand(tt.cmd, pathDirs, pathExt)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantAbs {
				assert.True(t, filepath.IsAbs(result), "expected absolute path, got: %s", result)
			}
		})
	}
}

func TestResolveCommand_AlreadyAbsolute(t *testing.T) {
	abs := `C:\Program Files\nodejs\npx.cmd`
	result, err := ResolveCommand(abs, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, abs, result)
}

func TestResolveCommand_PATHEXTOrdering(t *testing.T) {
	tmpDir := t.TempDir()

	// Create both .cmd and .exe — .cmd should win if listed first.
	for _, name := range []string{"npx.cmd", "npx.exe"} {
		f, err := os.Create(filepath.Join(tmpDir, name))
		require.NoError(t, err)
		f.Close()
	}

	result, err := ResolveCommand("npx", []string{tmpDir}, []string{".cmd", ".exe"})
	require.NoError(t, err)
	assert.Contains(t, result, "npx.cmd")
}

func TestSplitPathExt(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{".COM;.EXE;.BAT;.CMD", []string{".com", ".exe", ".bat", ".cmd"}},
		{"", nil},
		{".EXE", []string{".exe"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, splitPathExt(tt.input))
		})
	}
}
