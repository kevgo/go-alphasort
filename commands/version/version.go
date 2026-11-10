// Package version implements the "version" subcommand of alphasort.
package version

import (
	"fmt"
	"io"
	"runtime/debug"
	"strings"
)

// Determine provides the version of this alphasort binary.
// Release builds provide it via linker flags, all other builds via the build information embedded by the Go toolchain.
func Determine(ldflagsVersion string) string {
	if ldflagsVersion != "" {
		return strings.TrimPrefix(ldflagsVersion, "v")
	}
	info, hasInfo := debug.ReadBuildInfo()
	if !hasInfo || info.Main.Version == "" {
		return "(devel)"
	}
	return strings.TrimPrefix(info.Main.Version, "v")
}

// Run executes the "version" subcommand and returns the exit code.
func Run(args []string, version string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		fmt.Fprintf(stderr, "unexpected argument: %s\n", args[0])
		return 2
	}
	fmt.Fprintf(stdout, "alphasort %s\n", version)
	return 0
}
