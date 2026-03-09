package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFlags(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantFlags Flags
		wantPos   []string
	}{
		{
			name:      "no flags",
			args:      []string{"file.json"},
			wantFlags: Flags{},
			wantPos:   []string{"file.json"},
		},
		{
			name:      "yes flag",
			args:      []string{"file.json", "--yes"},
			wantFlags: Flags{Yes: true},
			wantPos:   []string{"file.json"},
		},
		{
			name:      "yes short flag",
			args:      []string{"-y", "file.json"},
			wantFlags: Flags{Yes: true},
			wantPos:   []string{"file.json"},
		},
		{
			name:      "multiple flags",
			args:      []string{"--yes", "--no-backup", "file.json"},
			wantFlags: Flags{Yes: true, NoBackup: true},
			wantPos:   []string{"file.json"},
		},
		{
			name:      "output flag",
			args:      []string{"-o", "out.json", "file.json"},
			wantFlags: Flags{Output: "out.json"},
			wantPos:   []string{"file.json"},
		},
		{
			name:      "version",
			args:      []string{"--version"},
			wantFlags: Flags{Version: true},
		},
		{
			name:      "help",
			args:      []string{"--help"},
			wantFlags: Flags{Help: true},
		},
		{
			name:      "double dash passthrough",
			args:      []string{"claude", "mcp", "add", "--", "npx", "-y", "@pkg"},
			wantFlags: Flags{},
			wantPos:   []string{"claude", "mcp", "add", "--", "npx", "-y", "@pkg"},
		},
		{
			name:      "unwrap and quiet",
			args:      []string{"--unwrap", "-q", `{"command":"cmd.exe"}`},
			wantFlags: Flags{Unwrap: true, Quiet: true},
			wantPos:   []string{`{"command":"cmd.exe"}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, pos := extractFlags(tt.args)
			assert.Equal(t, tt.wantFlags, flags)
			assert.Equal(t, tt.wantPos, pos)
		})
	}
}

func TestDetectMode(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want mode
	}{
		{"known provider claude", []string{"claude", "mcp", "add"}, modeCLI},
		{"known provider code", []string{"code", "--add-mcp", "{}"}, modeCLI},
		{"known provider q", []string{"q", "mcp", "add"}, modeCLI},
		{"inline JSON object", []string{`{"command":"npx"}`}, modeJSON},
		{"inline JSON array", []string{`[{"command":"npx"}]`}, modeJSON},
		{"json file extension", []string{"config.json"}, modeFile},
		{"config subcommand", []string{"config", "get"}, modeConfig},
		{"update subcommand", []string{"update"}, modeUpdate},
		{"unknown command (fallback)", []string{"somecmd", "arg1"}, modeCLI},
		{"empty args", nil, modeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectMode(tt.args, nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractFlagsStrict(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantFlags Flags
		wantPos   []string
	}{
		{
			name:      "flags before positional",
			args:      []string{"--dry-run", "-y", "claude", "mcp", "add"},
			wantFlags: Flags{DryRun: true, Yes: true},
			wantPos:   []string{"claude", "mcp", "add"},
		},
		{
			name:      "stops at first positional",
			args:      []string{"claude", "mcp", "add", "-y", "npx"},
			wantFlags: Flags{},
			wantPos:   []string{"claude", "mcp", "add", "-y", "npx"},
		},
		{
			name:      "preserves -y after provider",
			args:      []string{"claude", "mcp", "add", "srv", "npx", "-y", "@pkg"},
			wantFlags: Flags{},
			wantPos:   []string{"claude", "mcp", "add", "srv", "npx", "-y", "@pkg"},
		},
		{
			name:      "flag then -- preserved",
			args:      []string{"--quiet", "claude", "mcp", "add", "--", "npx", "-y"},
			wantFlags: Flags{Quiet: true},
			wantPos:   []string{"claude", "mcp", "add", "--", "npx", "-y"},
		},
		{
			name:      "no args",
			args:      nil,
			wantFlags: Flags{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flags, pos := extractFlagsStrict(tt.args)
			assert.Equal(t, tt.wantFlags, flags)
			assert.Equal(t, tt.wantPos, pos)
		})
	}
}

func TestContainsDashDash(t *testing.T) {
	assert.True(t, containsDashDash([]string{"a", "--", "b"}))
	assert.False(t, containsDashDash([]string{"a", "b"}))
	assert.False(t, containsDashDash(nil))
}

func TestRun_Version(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	SetVersion("1.0.0", "abc123", "2026-01-01")
	code := Run([]string{"--version"}, nil, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "1.0.0")
}

func TestRun_Help(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := Run([]string{"--help"}, nil, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "mcp2win")
}

func TestRun_InlineJSON(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := Run([]string{"--quiet", `{"command":"npx","args":["-y","@pkg"]}`}, nil, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout.String(), "cmd.exe")
}
