package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/operator-kit/mcp2win/internal/selfupdate"
)

func runUpdate(stdout, stderr io.Writer) int {
	if appVersion == "dev" {
		fmt.Fprintln(stderr, "Skipping update: running dev build")
		return 0
	}

	release, err := selfupdate.FetchLatestRelease()
	if err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	if release == nil {
		fmt.Fprintln(stdout, "No published release found.")
		return 0
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	if selfupdate.CompareVersions(appVersion, latest) >= 0 {
		fmt.Fprintf(stdout, "Already up to date (v%s).\n", appVersion)
		return 0
	}

	fmt.Fprintf(stderr, "Updating v%s → v%s\n", appVersion, latest)
	if err := selfupdate.Update(release, stderr); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Successfully updated to v%s.\n", latest)
	return 0
}
