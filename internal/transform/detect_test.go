package transform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectShape(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		shape   Shape
		key     string
		wantErr bool
	}{
		{
			name:  "full config with mcpServers",
			data:  map[string]any{"mcpServers": map[string]any{"s1": map[string]any{"command": "npx"}}},
			shape: ShapeFullConfig,
			key:   "mcpServers",
		},
		{
			name:  "full config with servers",
			data:  map[string]any{"servers": map[string]any{"s1": map[string]any{"command": "npx"}}},
			shape: ShapeFullConfig,
			key:   "servers",
		},
		{
			name:  "full config with context_servers (Zed)",
			data:  map[string]any{"context_servers": map[string]any{"s1": map[string]any{"command": "npx"}}},
			shape: ShapeFullConfig,
			key:   "context_servers",
		},
		{
			name:  "single server with command",
			data:  map[string]any{"command": "npx", "args": []any{"-y", "@pkg"}},
			shape: ShapeSingleServer,
		},
		{
			name:  "single server with url (HTTP)",
			data:  map[string]any{"url": "http://localhost:3000"},
			shape: ShapeSingleServer,
		},
		{
			name:  "single server with serverUrl",
			data:  map[string]any{"serverUrl": "http://localhost:3000"},
			shape: ShapeSingleServer,
		},
		{
			name:  "servers block (bare map of servers)",
			data:  map[string]any{"s1": map[string]any{"command": "npx"}, "s2": map[string]any{"command": "uvx"}},
			shape: ShapeServersBlock,
		},
		{
			name:  "servers block with HTTP servers",
			data:  map[string]any{"s1": map[string]any{"url": "http://localhost:3000"}},
			shape: ShapeServersBlock,
		},
		{
			name:    "unrecognized structure",
			data:    map[string]any{"foo": "bar", "baz": 42},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shape, key, err := DetectShape(tt.data)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.shape, shape)
			assert.Equal(t, tt.key, key)
		})
	}
}
