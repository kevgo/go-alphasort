package fix_test

import (
	"io/fs"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/kevgo/go-alphasort/commands/fix"
)

func TestFind(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"main.go":                 {},
		"main_test.go":            {},
		"README.md":               {},
		".hidden.go":              {},
		"dir.go/file.txt":         {},
		"pkg/lib.go":              {},
		"pkg/sub/deep.go":         {},
		"pkg/notes.txt":           {},
		"pkg/vendor/nested.go":    {},
		"vendor/dep/dep.go":       {},
		"testdata/fixture.go":     {},
		".git/hooks/hook.go":      {},
		"_build/build.go":         {},
		"pkg/_internal/x.go":      {},
		"pkg/testdata/golden.go":  {},
		"pkg/sub/.cache/cache.go": {},
	}
	have, err := fix.Find(fsys)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"main.go",
		"main_test.go",
		"pkg/lib.go",
		"pkg/sub/deep.go",
	}
	if !slices.Equal(have, want) {
		t.Errorf("mismatch\nhave: %q\nwant: %q", have, want)
	}
}

func TestFindEmpty(t *testing.T) {
	t.Parallel()
	have, err := fix.Find(fstest.MapFS{})
	if err != nil {
		t.Fatal(err)
	}
	if len(have) != 0 {
		t.Errorf("expected no files, got %q", have)
	}
}

func TestIsGoFile(t *testing.T) {
	t.Parallel()
	fsys := fstest.MapFS{
		"main.go":       {},
		"main_test.go":  {},
		".hidden.go":    {},
		"notes.txt":     {},
		"go":            {},
		"file.go.txt":   {},
		"dir.go/x.txt":  {},
		"plaindir/x.go": {},
	}
	tests := map[string]bool{
		"main.go":      true,
		"main_test.go": true,
		".hidden.go":   false,
		"notes.txt":    false,
		"go":           false,
		"file.go.txt":  false,
		"dir.go":       false,
		"plaindir":     false,
	}
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(tests) {
		t.Fatalf("expected %d entries, got %d", len(tests), len(entries))
	}
	for _, entry := range entries {
		if have, want := fix.IsGoFile(entry), tests[entry.Name()]; have != want {
			t.Errorf("IsGoFile(%q): have %v, want %v", entry.Name(), have, want)
		}
	}
}
