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
		".git/hooks/hook.go":      {},
		".hidden.go":              {},
		"README.md":               {},
		"_build/build.go":         {},
		"dir.go/file.txt":         {},
		"main.go":                 {},
		"main_test.go":            {},
		"pkg/_internal/x.go":      {},
		"pkg/lib.go":              {},
		"pkg/notes.txt":           {},
		"pkg/sub/.cache/cache.go": {},
		"pkg/sub/deep.go":         {},
		"pkg/testdata/golden.go":  {},
		"pkg/vendor/nested.go":    {},
		"testdata/fixture.go":     {},
		"vendor/dep/dep.go":       {},
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
		".hidden.go":    {},
		"dir.go/x.txt":  {},
		"file.go.txt":   {},
		"go":            {},
		"main.go":       {},
		"main_test.go":  {},
		"notes.txt":     {},
		"plaindir/x.go": {},
	}
	tests := map[string]bool{
		".hidden.go":   false,
		"dir.go":       false,
		"file.go.txt":  false,
		"go":           false,
		"main.go":      true,
		"main_test.go": true,
		"notes.txt":    false,
		"plaindir":     false,
	}
	entries, err := fs.ReadDir(fsys, ".")
	must.NoError(t, err)
	must.SliceLen(t, len(tests), entries)
	for _, give := range entries {
		have := fix.IsGoFile(give)
		want := tests[give.Name()]
		must.Eq(t, want, have, must.Sprintf("IsGoFile(%q)", give.Name()))
	}
}
