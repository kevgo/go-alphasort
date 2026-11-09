package fix_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/kevgo/go-alphasort/commands/fix"
	"github.com/shoenig/test/must"
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
	must.NoError(t, err)
	want := []string{
		"main.go",
		"main_test.go",
		"pkg/lib.go",
		"pkg/sub/deep.go",
	}
	must.Eq(t, want, have)
}

func TestFindEmpty(t *testing.T) {
	t.Parallel()
	have, err := fix.Find(fstest.MapFS{})
	must.NoError(t, err)
	must.SliceEmpty(t, have)
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
	must.NoError(t, err)
	must.SliceLen(t, len(tests), entries)
	for _, entry := range entries {
		must.Eq(t, tests[entry.Name()], fix.IsGoFile(entry), must.Sprintf("IsGoFile(%q)", entry.Name()))
	}
}
