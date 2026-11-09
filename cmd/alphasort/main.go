// Alphasort sorts the free-standing functions in Go source files alphabetically.
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevgo/go-alphasort/fixer"
)

// isGoFile indicates whether the given directory entry is a Go source file.
func isGoFile(entry fs.DirEntry) bool {
	name := entry.Name()
	return !entry.IsDir() && strings.HasSuffix(name, ".go") && !strings.HasPrefix(name, ".")
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// printUsage prints the help text of this tool.
func printUsage(w io.Writer) {
	fmt.Fprint(w, `usage: alphasort <command> [path ...]

Sorts free-standing Go functions alphabetically.

Commands:
  check  list the files whose functions are not sorted
  fix    sort the functions in the given files
  help   show this help

Directories are processed recursively.
Without paths, reads Go source code from STDIN.
In this case, "fix" writes the sorted code to STDOUT.
`)
}

// processFile checks or fixes the given file.
func processFile(path string, cmd command, stdout io.Writer) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	sorted, err := fixer.SortFuncs(path, src)
	if err != nil {
		return err
	}
	if bytes.Equal(src, sorted) {
		return nil
	}
	switch cmd {
	case commandCheck:
		_, err = fmt.Fprintln(stdout, path)
		return err
	case commandFix:
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		return os.WriteFile(path, sorted, info.Mode().Perm())
	}
	return nil
}

// processPath checks or fixes the given file, or all Go files in the given directory tree.
func processPath(root string, cmd command, stdout io.Writer) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return processFile(root, cmd, stdout)
	}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && path != root && skipDir(entry.Name()) {
			return filepath.SkipDir
		}
		if !isGoFile(entry) {
			return nil
		}
		return processFile(path, cmd, stdout)
	})
}

// processStdin checks or fixes the Go source code provided via STDIN.
func processStdin(cmd command, stdin io.Reader, stdout io.Writer) error {
	src, err := io.ReadAll(stdin)
	if err != nil {
		return err
	}
	sorted, err := fixer.SortFuncs("<standard input>", src)
	if err != nil {
		return err
	}
	switch cmd {
	case commandCheck:
		if !bytes.Equal(src, sorted) {
			_, err = fmt.Fprintln(stdout, "<standard input>")
		}
	case commandFix:
		_, err = stdout.Write(sorted)
	}
	return err
}

// run executes this tool with the given CLI arguments and returns the exit code.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	var cmd command
	switch args[0] {
	case "check":
		cmd = commandCheck
	case "fix":
		cmd = commandFix
	case "help", "-h", "-help", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printUsage(stderr)
		return 2
	}
	paths := args[1:]
	for _, path := range paths {
		if strings.HasPrefix(path, "-") {
			fmt.Fprintf(stderr, "unknown flag: %s\n", path)
			return 2
		}
	}
	if len(paths) == 0 {
		if err := processStdin(cmd, stdin, stdout); err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		return 0
	}
	exitCode := 0
	for _, path := range paths {
		if err := processPath(path, cmd, stdout); err != nil {
			fmt.Fprintln(stderr, err)
			exitCode = 2
		}
	}
	return exitCode
}

// skipDir indicates whether the directory with the given name should not be processed,
// following the conventions of the Go tool.
func skipDir(name string) bool {
	return name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// command is a subcommand of this tool.
type command int

const (
	commandCheck command = iota // list files that aren't sorted
	commandFix                  // sort files
)
