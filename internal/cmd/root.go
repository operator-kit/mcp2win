package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/operator-kit/mcp2win/internal/color"
)

var (
	appVersion = "dev"
	appCommit  = "none"
	appDate    = "unknown"
)

// SetVersion sets version info from ldflags.
func SetVersion(version, commit, date string) {
	appVersion = version
	appCommit = commit
	appDate = date
}

// Flags holds parsed CLI flags.
type Flags struct {
	Write    bool
	NoBackup bool
	Output   string
	DryRun   bool
	Quiet    bool
	Unwrap   bool
	Resolve  bool
	NoColor  bool
	Version  bool
	Help     bool
}

// Known CLI providers for Mode 1.
var knownProviders = map[string]bool{
	"claude": true, "gemini": true, "code": true,
	"qchat": true, "q": true,
}

// Run is the main entry point. Returns exit code.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	color.DisableIfFlag(args)
	flags, positional := extractFlags(args)

	if flags.NoColor {
		color.Disabled = true
	}

	if flags.Version {
		printVersion(stdout)
		return 0
	}

	if flags.Help {
		printUsage(stdout)
		return 0
	}

	// Mode detection.
	mode := detectMode(positional, stdin)

	switch mode {
	case modeCLI:
		return runCLI(positional, flags, stdout, stderr)
	case modeJSON:
		return runJSON(positional, flags, stdin, stdout, stderr)
	case modeFile:
		return runFile(positional, flags, stdout, stderr)
	default:
		printUsage(stderr)
		return 1
	}
}

type mode int

const (
	modeUnknown mode = iota
	modeCLI
	modeJSON
	modeFile
)

func detectMode(positional []string, stdin io.Reader) mode {
	// Check stdin — pipe or redirected file.
	if f, ok := stdin.(*os.File); ok {
		if stat, err := f.Stat(); err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
			return modeJSON
		}
	}

	if len(positional) == 0 {
		return modeUnknown
	}

	first := positional[0]

	// Known provider → Mode 1.
	if knownProviders[strings.ToLower(first)] {
		return modeCLI
	}

	// Starts with { or [ → inline JSON (Mode 2).
	trimmed := strings.TrimSpace(first)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		return modeJSON
	}

	// Existing file → Mode 3.
	if _, err := os.Stat(first); err == nil {
		return modeFile
	}

	// Ends with .json → Mode 3 (will error if not found).
	if strings.HasSuffix(strings.ToLower(first), ".json") {
		return modeFile
	}

	// Fallback → Mode 1 (unknown provider).
	return modeCLI
}

// extractFlags pre-scans args and separates flags from positional args.
// This allows flags to appear anywhere (e.g., `mcp2win file.json --write`).
func extractFlags(args []string) (Flags, []string) {
	var f Flags
	var positional []string

	i := 0
	for i < len(args) {
		a := args[i]
		switch a {
		case "--write":
			f.Write = true
		case "--no-backup":
			f.NoBackup = true
		case "-o":
			if i+1 < len(args) {
				i++
				f.Output = args[i]
			}
		case "--dry-run":
			f.DryRun = true
		case "--quiet", "-q":
			f.Quiet = true
		case "--unwrap":
			f.Unwrap = true
		case "--resolve":
			f.Resolve = true
		case "--no-color":
			f.NoColor = true
		case "--version", "-v":
			f.Version = true
		case "--help", "-h":
			f.Help = true
		case "--":
			// Everything after -- is positional (pass-through for Mode 1).
			positional = append(positional, args[i:]...)
			i = len(args)
			continue
		default:
			positional = append(positional, a)
		}
		i++
	}

	return f, positional
}
