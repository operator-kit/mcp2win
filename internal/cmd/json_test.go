package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunJSON_InlineSingleServer(t *testing.T) {
	input := `{"command":"npx","args":["-y","@pkg"]}`
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	code := runJSON([]string{input}, Flags{Quiet: true}, nil, stdout, stderr)
	assert.Equal(t, 0, code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &out))
	assert.Equal(t, "cmd.exe", out["command"])
	args := out["args"].([]any)
	assert.Equal(t, "/c", args[0])
	assert.Equal(t, "npx", args[1])
}

func TestRunJSON_InlineFullConfig(t *testing.T) {
	input := `{"mcpServers":{"s1":{"command":"npx","args":["-y","@pkg"]},"s2":{"url":"http://localhost"}}}`
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	code := runJSON([]string{input}, Flags{Quiet: true}, nil, stdout, stderr)
	assert.Equal(t, 0, code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &out))
	servers := out["mcpServers"].(map[string]any)
	s1 := servers["s1"].(map[string]any)
	assert.Equal(t, "cmd.exe", s1["command"])
}

func TestRunJSON_Stdin(t *testing.T) {
	input := `{"command":"uvx","args":["mcp-server-fetch"]}`
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	stdin := strings.NewReader(input)

	code := runJSON(nil, Flags{Quiet: true}, stdin, stdout, stderr)
	assert.Equal(t, 0, code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &out))
	assert.Equal(t, "cmd.exe", out["command"])
}

func TestRunJSON_ServersBlock(t *testing.T) {
	input := `{"s1":{"command":"npx","args":["-y","@pkg"]},"s2":{"command":"uvx","args":["fetch"]}}`
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	code := runJSON([]string{input}, Flags{Quiet: true}, nil, stdout, stderr)
	assert.Equal(t, 0, code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &out))
	s1 := out["s1"].(map[string]any)
	assert.Equal(t, "cmd.exe", s1["command"])
}

func TestRunJSON_Unwrap(t *testing.T) {
	input := `{"command":"cmd.exe","args":["/c","npx","-y","@pkg"]}`
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	code := runJSON([]string{input}, Flags{Quiet: true, Unwrap: true}, nil, stdout, stderr)
	assert.Equal(t, 0, code)

	var out map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &out))
	assert.Equal(t, "npx", out["command"])
	args := out["args"].([]any)
	assert.Equal(t, "-y", args[0])
}

func TestRunJSON_InvalidJSON(t *testing.T) {
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	code := runJSON([]string{"not json"}, Flags{}, nil, stdout, stderr)
	assert.Equal(t, 1, code)
	assert.Contains(t, stderr.String(), "invalid JSON")
}

func TestRunJSON_Preview(t *testing.T) {
	input := `{"command":"npx","args":["-y","@pkg"]}`
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}

	code := runJSON([]string{input}, Flags{}, nil, stdout, stderr)
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr.String(), "Changes: 1")
}
