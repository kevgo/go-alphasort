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

// findGoFiles provides all Go files in the current directory and its subdirectories,
// skipping the directories that the Go tool ignores.
func findGoFiles() ([]string, error) {
	result := []string{}
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && path != "." && skipDir(entry.Name()) {
			return filepath.SkipDir
		}
		if isGoFile(entry) {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}

// fixFile sorts the functions in the given file and writes the result back to disk.
func fixFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory: please provide files, or no arguments to fix all files", path)
	}
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
	return os.WriteFile(path, sorted, info.Mode().Perm())
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
	fmt.Fprint(w, `usage: alphasort <command> [file ...]

Sorts free-standing Go functions alphabetically.

Commands:
  fix   sort the functions in the given files,
        or in all Go files in the current directory and its subdirectories
  help  show this help
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
	files := args[1:]
	for _, file := range files {
		if strings.HasPrefix(file, "-") {
			fmt.Fprintf(stderr, "unknown flag: %s\n", file)
			return 2
		}
	}
	if len(files) == 0 {
		var err error
		files, err = findGoFiles()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
	}
	exitCode := 0
	for _, file := range files {
		if err := fixFile(file); err != nil {
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
