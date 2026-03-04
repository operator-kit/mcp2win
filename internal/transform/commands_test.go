package transform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNeedsWrapping(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"npx", true},
		{"NPX", true},
		{"npx.exe", true},
		{"pnpx", true},
		{"bunx", true},
		{"uvx", true},
		{"yarn", true},
		{"tsx", true},
		{"prettier", true},
		{"node", false},
		{"python", false},
		{"cmd", false},
		{"unknown-tool", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			assert.Equal(t, tt.want, NeedsWrapping(tt.cmd))
		})
	}
}

func TestIsNativeExe(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"node", true},
		{"NODE", true},
		{"node.exe", true},
		{"python", true},
		{"python3", true},
		{"deno", true},
		{"bun", true},
		{"cmd", true},
		{"npx", false},
		{"unknown", false},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			assert.Equal(t, tt.want, IsNativeExe(tt.cmd))
		})
	}
}

func TestIsAlreadyWrapped(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"cmd", true},
		{"CMD", true},
		{"cmd.exe", true},
		{"CMD.EXE", true},
		{"npx", false},
		{"node", false},
	}
	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			assert.Equal(t, tt.want, IsAlreadyWrapped(tt.cmd))
		})
	}
}
