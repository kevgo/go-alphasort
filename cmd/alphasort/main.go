// Alphasort sorts the free-standing functions in Go source files alphabetically.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kevgo/go-alphasort/commands/fix"
	"github.com/kevgo/go-alphasort/commands/help"
	"github.com/kevgo/go-alphasort/commands/version"
)

// buildVersion contains the version of this binary.
// Release builds set it via linker flags.
var buildVersion = ""

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes this tool with the given CLI arguments and returns the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		help.Print(stderr)
		return 2
	}
	switch args[0] {
	case "fix":
		return fix.Run(args[1:], stderr)
	case "help", "-h", "-help", "--help":
		return help.Run(stdout)
	case "version", "-version", "--version":
		return version.Run(args[1:], version.Determine(buildVersion), stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		help.Print(stderr)
		return 2
	}
}
