package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/operator-kit/mcp2win/internal/config"
)

func runConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printConfigUsage(stderr)
		return 1
	}

	switch args[0] {
	case "get":
		return runConfigGet(args[1:], stdout, stderr)
	case "set":
		return runConfigSet(args[1:], stdout, stderr)
	case "path":
		fmt.Fprintln(stdout, config.DefaultPath())
		return 0
	case "reset":
		return runConfigReset(stdout, stderr)
	default:
		fmt.Fprintf(stderr, "Error: unknown config subcommand %q\n", args[0])
		printConfigUsage(stderr)
		return 1
	}
}

func runConfigGet(args []string, stdout, stderr io.Writer) int {
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	if len(args) == 0 {
		fmt.Fprintf(stdout, "always_exec_cli: %t\n", cfg.AlwaysExecCLI)
		fmt.Fprintf(stdout, "always_write_file: %t\n", cfg.AlwaysWriteFile)
		return 0
	}

	switch args[0] {
	case "always_exec_cli":
		fmt.Fprintf(stdout, "%t\n", cfg.AlwaysExecCLI)
	case "always_write_file":
		fmt.Fprintf(stdout, "%t\n", cfg.AlwaysWriteFile)
	default:
		fmt.Fprintf(stderr, "Error: unknown config key %q\n", args[0])
		fmt.Fprintln(stderr, "Valid keys: always_exec_cli, always_write_file")
		return 1
	}
	return 0
}

func runConfigSet(args []string, stdout, stderr io.Writer) int {
	if len(args) < 2 {
		fmt.Fprintln(stderr, "Usage: mcp2win config set <key> <value>")
		fmt.Fprintln(stderr, "Valid keys: always_exec_cli, always_write_file")
		fmt.Fprintln(stderr, "Valid values: true, false")
		return 1
	}

	key, valStr := args[0], args[1]
	val, err := strconv.ParseBool(valStr)
	if err != nil {
		fmt.Fprintf(stderr, "Error: invalid value %q (expected true or false)\n", valStr)
		return 1
	}

	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	switch key {
	case "always_exec_cli":
		cfg.AlwaysExecCLI = val
	case "always_write_file":
		cfg.AlwaysWriteFile = val
	default:
		fmt.Fprintf(stderr, "Error: unknown config key %q\n", key)
		fmt.Fprintln(stderr, "Valid keys: always_exec_cli, always_write_file")
		return 1
	}

	if err := config.Save("", cfg); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Config saved to %s\n", config.DefaultPath())
	return 0
}

func runConfigReset(stdout, stderr io.Writer) int {
	path := config.DefaultPath()
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(stdout, "No config file to reset")
			return 0
		}
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Config reset (removed %s)\n", path)
	return 0
}

func printConfigUsage(w io.Writer) {
	fmt.Fprintln(w, `Usage: mcp2win config <command>

Commands:
  get [key]          Show config values (all or specific key)
  set <key> <value>  Set a config value
  path               Show config file location
  reset              Remove config file (reset all preferences)

Keys:
  always_exec_cli    Skip confirmation when executing CLI commands
  always_write_file  Skip confirmation when writing config files`)
}
