package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/operator-kit/mcp2win/internal/color"
	"github.com/operator-kit/mcp2win/internal/config"
)

// confirmAction prompts the user for confirmation. Returns true if the action
// should proceed. If the user answers "always", saves the preference to config.
//
// Skips the prompt and returns true if:
//   - flags.Yes is set (--yes / -y)
//   - the relevant config preference is already set
//   - stdin is not a terminal (non-interactive, requires --yes)
func confirmAction(prompt string, configKey string, flags Flags, cfg *config.Config, stderr io.Writer) bool {
	// --yes flag: skip prompt.
	if flags.Yes {
		return true
	}

	// Config preference: skip prompt.
	if configKey == "always_exec_cli" && cfg.AlwaysExecCLI {
		return true
	}
	if configKey == "always_write_file" && cfg.AlwaysWriteFile {
		return true
	}

	// Non-interactive stdin: can't prompt.
	if !isTerminal(os.Stdin) {
		fmt.Fprintf(stderr, "%s non-interactive mode, use --yes to confirm\n", color.Yellow("Note:"))
		return false
	}

	fmt.Fprintf(stderr, "%s [y]es / [n]o / [a]lways: ", prompt)

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	switch answer {
	case "y", "yes":
		return true
	case "a", "always":
		if configKey == "always_exec_cli" {
			cfg.AlwaysExecCLI = true
		}
		if configKey == "always_write_file" {
			cfg.AlwaysWriteFile = true
		}
		if err := config.Save("", cfg); err != nil {
			fmt.Fprintf(stderr, "%s could not save preference: %v\n", color.Yellow("Warning:"), err)
		} else {
			fmt.Fprintf(stderr, "%s Reset with: mcp2win config reset\n", color.Dim("Preference saved."))
		}
		return true
	default:
		return false
	}
}

func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
