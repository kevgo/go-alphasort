// Package fix implements the "fix" subcommand of alphasort.
package fix

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Run executes the "fix" subcommand with the given arguments and returns the exit code.
// Without arguments, it fixes all Go files in the current directory and its subdirectories.
func Run(args []string, stderr io.Writer) int {
	files := args
	for _, file := range files {
		if strings.HasPrefix(file, "-") {
			fmt.Fprintf(stderr, "unknown flag: %s\n", file)
			return 2
		}
	}
	if len(files) == 0 {
		found, err := Find(os.DirFS("."))
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

// fixFile sorts the given file and writes the result back to disk.
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
	sorted, err := Sort(path, src)
	if err != nil {
		return err
	}
	if bytes.Equal(src, sorted) {
		return nil
	}
	return os.WriteFile(path, sorted, info.Mode().Perm())
}
