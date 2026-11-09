// go-alphasort sorts the free-standing functions in Go source files alphabetically.
package main

import (
	"bytes"
	"errors"
	"flag"
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

// processFile sorts the functions in the given file according to the given options.
func processFile(path string, opts options, stdout io.Writer) error {
	src, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	sorted, err := fixer.SortFuncs(path, src)
	if err != nil {
		return err
	}
	changed := !bytes.Equal(src, sorted)
	if changed && opts.list {
		fmt.Fprintln(stdout, path)
	}
	if opts.write {
		if !changed {
			return nil
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		return os.WriteFile(path, sorted, info.Mode().Perm())
	}
	if !opts.list {
		_, err = stdout.Write(sorted)
	}
	return err
}

// processPath sorts the given file, or all Go files in the given directory tree.
func processPath(root string, opts options, stdout io.Writer) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return processFile(root, opts, stdout)
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
		return processFile(path, opts, stdout)
	})
}

// processStdin sorts the Go source code provided via STDIN and prints the result to STDOUT.
func processStdin(opts options, stdin io.Reader, stdout io.Writer) error {
	if opts.write {
		return errors.New("cannot use -w with standard input")
	}
	src, err := io.ReadAll(stdin)
	if err != nil {
		return err
	}
	sorted, err := fixer.SortFuncs("<standard input>", src)
	if err != nil {
		return err
	}
	if opts.list {
		if !bytes.Equal(src, sorted) {
			fmt.Fprintln(stdout, "<standard input>")
		}
		return nil
	}
	_, err = stdout.Write(sorted)
	return err
}

// run executes this tool with the given CLI arguments and returns the exit code.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("go-alphasort", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var opts options
	flags.BoolVar(&opts.list, "l", false, "list files whose functions are not sorted alphabetically")
	flags.BoolVar(&opts.write, "w", false, "write the result to the source files instead of STDOUT")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: go-alphasort [flags] [path ...]")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Sorts free-standing Go functions alphabetically.")
		fmt.Fprintln(stderr, "Without paths, reads Go source code from STDIN.")
		fmt.Fprintln(stderr)
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() == 0 {
		if err := processStdin(opts, stdin, stdout); err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		return 0
	}
	exitCode := 0
	for _, path := range flags.Args() {
		if err := processPath(path, opts, stdout); err != nil {
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

type options struct {
	list  bool
	write bool
}
