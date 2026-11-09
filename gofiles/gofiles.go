// Package gofiles finds Go source files.
package gofiles

import (
	"io/fs"
	"strings"
)

// Find provides the paths of all Go files in the given file system,
// skipping the directories that the Go tool ignores:
// vendor, testdata, and directories whose name starts with "." or "_".
// The paths are slash-separated and relative to the root of the file system.
func Find(fsys fs.FS) ([]string, error) {
	result := []string{}
	err := fs.WalkDir(fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && path != "." && skipDir(entry.Name()) {
			return fs.SkipDir
		}
		if IsGoFile(entry) {
			result = append(result, path)
		}
		return nil
	})
	return result, err
}

// IsGoFile indicates whether the given directory entry is a Go source file.
func IsGoFile(entry fs.DirEntry) bool {
	name := entry.Name()
	return !entry.IsDir() && strings.HasSuffix(name, ".go") && !strings.HasPrefix(name, ".")
}

// skipDir indicates whether the directory with the given name should not be searched,
// following the conventions of the Go tool.
func skipDir(name string) bool {
	return name == "vendor" || name == "testdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}
