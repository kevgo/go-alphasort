// Alphasort sorts the free-standing functions in Go source files alphabetically.
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kevgo/go-alphasort/fixer"
	"github.com/kevgo/go-alphasort/gofiles"
)

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
		found, err := gofiles.Find(os.DirFS("."))
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		for _, file := range found {
			files = append(files, filepath.FromSlash(file))
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
