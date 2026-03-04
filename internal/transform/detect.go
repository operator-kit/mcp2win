package transform

import "fmt"

// Shape describes the structure of input JSON.
type Shape int

const (
	ShapeUnknown      Shape = iota
	ShapeSingleServer       // {"command":"npx","args":[...]}
	ShapeServersBlock       // {"server-name": {"command":...}, ...}
	ShapeFullConfig         // {"mcpServers": {...}} or {"servers": {...}} etc.
)

func (s Shape) String() string {
	switch s {
	case ShapeSingleServer:
		return "SINGLE_SERVER"
	case ShapeServersBlock:
		return "SERVERS_BLOCK"
	case ShapeFullConfig:
		return "FULL_CONFIG"
	default:
		return "UNKNOWN"
	}
}

// Known top-level keys that wrap server blocks.
var configKeys = []string{
	"mcpServers",
	"servers",
	"context_servers",
}

// DetectShape analyzes JSON data and returns the shape, the key containing servers (if any), and an error.
func DetectShape(data map[string]any) (Shape, string, error) {
	// Check for known config wrapper keys.
	for _, key := range configKeys {
		if val, ok := data[key]; ok {
			if _, isMap := val.(map[string]any); isMap {
				return ShapeFullConfig, key, nil
			}
		}
	}

	// Check for single server indicators.
	if _, hasCmd := data["command"]; hasCmd {
		return ShapeSingleServer, "", nil
	}
	if _, hasURL := data["url"]; hasURL {
		return ShapeSingleServer, "", nil
	}
	if _, hasURL := data["serverUrl"]; hasURL {
		return ShapeSingleServer, "", nil
	}

	// Check if it looks like a servers block (all values are objects with command/url/serverUrl).
	if len(data) > 0 && allServers(data) {
		return ShapeServersBlock, "", nil
	}

	return ShapeUnknown, "", fmt.Errorf("unrecognized JSON structure: no mcpServers/servers/context_servers key, no command/url field, and values don't look like server definitions")
}

// allServers returns true if every value in data looks like a server definition.
func allServers(data map[string]any) bool {
	for _, v := range data {
		m, ok := v.(map[string]any)
		if !ok {
			return false
		}
		if _, hasCmd := m["command"]; hasCmd {
			continue
		}
		if _, hasURL := m["url"]; hasURL {
			continue
		}
		if _, hasURL := m["serverUrl"]; hasURL {
			continue
		}
		return false
	}
	return true
}
