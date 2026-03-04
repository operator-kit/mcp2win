package main

import (
	"os"

	"github.com/operator-kit/mcp2win/internal/cmd"
)

// Set via ldflags at build time.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date)
	os.Exit(cmd.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
