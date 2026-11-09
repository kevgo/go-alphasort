// Package help implements the "help" subcommand of alphasort.
package help

import (
	"fmt"
	"io"
)

// Print writes the usage information of alphasort to the given writer.
func Print(w io.Writer) {
	fmt.Fprint(w, `usage: alphasort <command> [file ...]

Sorts free-standing Go functions alphabetically.

Commands:
  fix   sort the functions in the given files,
        or in all Go files in the current directory and its subdirectories
  help  show this help
`)
}

// Run executes the "help" subcommand and returns the exit code.
func Run(stdout io.Writer) int {
	Print(stdout)
	return 0
}
