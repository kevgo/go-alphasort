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

// fixFile sorts the functions in the given file and writes the result back to disk.
func fixFile(path string) error {
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
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	return os.WriteFile(path, sorted, info.Mode().Perm())
}

// fixPath sorts the functions in the given file, or in all Go files in the given directory tree.
func fixPath(root string) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fixFile(root)
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
		return fixFile(path)
	})
}

// isGoFile indicates whether the given directory entry is a Go source file.
func isGoFile(entry fs.DirEntry) bool {
	name := entry.Name()
	return !entry.IsDir() && strings.HasSuffix(name, ".go") && !strings.HasPrefix(name, ".")
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// printUsage prints the help text of this tool.
func printUsage(w io.Writer) {
	fmt.Fprint(w, `usage: alphasort <command> [path ...]

Sorts free-standing Go functions alphabetically.

Commands:
  fix   sort the functions in the given files and directories
  help  show this help

Directories are processed recursively.
`)
}

// run executes this tool with the given CLI arguments and returns the exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	switch args[0] {
	case "fix":
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
		fmt.Fprintln(stderr, "please provide the files or directories to fix")
		return 2
	}
	exitCode := 0
	for _, path := range paths {
		if err := fixPath(path); err != nil {
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
