package fix_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kevgo/go-alphasort/commands/fix"
	"github.com/shoenig/test/must"
)

func TestSortFields(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		give string
		want string
	}{
		"struct definition": {
			give: `package a

type T struct {
	zulu  int
	Bravo string
	alpha bool
	Alpha bool
}
`,
			want: `package a

type T struct {
	Alpha bool
	Bravo string
	alpha bool
	zulu  int
}
`,
		},
		"comments move with their field": {
			give: `package a

type T struct {
	// zulu is Z.
	zulu int // trailing zulu

	// alpha is A.
	alpha int /* trailing alpha */
}
`,
			want: `package a

type T struct {
	// alpha is A.
	alpha int /* trailing alpha */

	// zulu is Z.
	zulu int // trailing zulu
}
`,
		},
		"floating comments stay in place": {
			give: `package a

type T struct {
	// section 1

	zulu int

	// section 2

	alpha int
	// end
}
`,
			want: `package a

type T struct {
	// section 1

	alpha int

	// section 2

	zulu int
	// end
}
`,
		},
		"realigns the fields": {
			give: `package a

type T struct {
	zulu    int
	charlie string

	alpha bool
}
`,
			want: `package a

type T struct {
	alpha   bool
	charlie string

	zulu int
}
`,
		},
		"single-line struct": {
			give: `package a

type T struct{ b int; a string }
`,
			want: `package a

type T struct{ a string; b int }
`,
		},
		"names in a multi-name field": {
			give: `package a

type T struct {
	z    int
	c, b string
}
`,
			want: `package a

type T struct {
	b, c string
	z    int
}
`,
		},
		"embedded fields sort by their type name": {
			give: `package a

import "sync"

type T struct {
	name string
	*sync.Mutex
	Base
}
`,
			want: `package a

import "sync"

type T struct {
	Base
	*sync.Mutex
	name string
}
`,
		},
		"tags move with their field": {
			give: "package a\n\ntype T struct {\n\tZ int    `json:\"z\"`\n\tA string `json:\"a\"`\n}\n",
			want: "package a\n\ntype T struct {\n\tA string `json:\"a\"`\n\tZ int    `json:\"z\"`\n}\n",
		},
		"nested and anonymous structs": {
			give: `package a

type T struct {
	zulu struct {
		y int
		x int
	}
	alpha int
}

func f(cfg struct{ b, a int }) {
	var v struct {
		d int
		c int
	}
	_ = v
}
`,
			want: `package a

type T struct {
	alpha int
	zulu  struct {
		x int
		y int
	}
}

func f(cfg struct{ a, b int }) {
	var v struct {
		c int
		d int
	}
	_ = v
}
`,
		},
		"structs with blank fields keep their order": {
			give: `package a

type T struct {
	b int32
	_ [4]byte
	a int64
}
`,
			want: `package a

type T struct {
	b int32
	_ [4]byte
	a int64
}
`,
		},
		"keyed instantiations": {
			give: `package a

var x = T{zulu: 1, Bravo: 2, alpha: 3}

var y = &T{
	zulu: 1, // trailing zulu
	// doc of alpha
	alpha: 3,
}

type T struct {
	Bravo int
	alpha int
	zulu  int
}
`,
			want: `package a

var x = T{Bravo: 2, alpha: 3, zulu: 1}

var y = &T{
	// doc of alpha
	alpha: 3,
	zulu:  1, // trailing zulu
}

type T struct {
	Bravo int
	alpha int
	zulu  int
}
`,
		},
		"nested and elided instantiations": {
			give: `package a

import "other"

var x = []Outer{
	{inner: Inner{d: 1, c: 2}, b: 3},
	{b: 4},
}

var y = map[string]*Inner{"one": {d: 1, c: 2}}

var z = other.Config{Zulu: 1, Alpha: 2}

var w = struct{ b, a int }{b: 1, a: 2}
`,
			want: `package a

import "other"

var x = []Outer{
	{b: 3, inner: Inner{c: 2, d: 1}},
	{b: 4},
}

var y = map[string]*Inner{"one": {c: 2, d: 1}}

var z = other.Config{Alpha: 2, Zulu: 1}

var w = struct{ a, b int }{a: 2, b: 1}
`,
		},
		"map and array literals keep their order": {
			give: `package a

const (
	z = iota
	a
)

var m = map[int]string{z: "z", a: "a"}

var s = [...]string{z: "z", a: "a"}

type Names map[int]string

var n = Names{z: "z", a: "a"}

type Tables []map[int]string

var l = Tables{
	{z: "z", a: "a"},
}
`,
			want: `package a

const (
	z = iota
	a
)

var m = map[int]string{z: "z", a: "a"}

var s = [...]string{z: "z", a: "a"}

type Names map[int]string

var n = Names{z: "z", a: "a"}

type Tables []map[int]string

var l = Tables{
	{z: "z", a: "a"},
}
`,
		},
		"instantiations with several side effects keep their order": {
			give: `package a

var x = T{b: next(), a: next()}

var y = T{b: <-ch, a: <-ch}

var z = T{b: next(), a: len(s), c: int64(1), d: MyInt(2), e: []byte("x"), f: func() int { return next() }}

type MyInt int
`,
			want: `package a

var x = T{b: next(), a: next()}

var y = T{b: <-ch, a: <-ch}

var z = T{a: len(s), b: next(), c: int64(1), d: MyInt(2), e: []byte("x"), f: func() int { return next() }}

type MyInt int
`,
		},
		"structs instantiated with positional fields get sorted as well": {
			give: `package a

type Point struct {
	y int
	x int
}

var p = Point{1, 2}

var tests = []struct {
	want int
	give int
}{
	{1, 2},
}
`,
			want: `package a

type Point struct {
	x int
	y int
}

var p = Point{1, 2}

var tests = []struct {
	give int
	want int
}{
	{1, 2},
}
`,
		},
		"struct fields and functions together": {
			give: `package a

func zulu() T { return T{b: 1, a: 2} }

func alpha() {}

type T struct {
	b int
	a int
}
`,
			want: `package a

func alpha() {}

func zulu() T { return T{a: 2, b: 1} }

type T struct {
	a int
	b int
}
`,
		},
		"not gofmt-formatted input does not get formatted": {
			give: `package a

type T struct {
  b   int
  a string
}
`,
			want: `package a

type T struct {
  a string
  b   int
}
`,
		},
		"CRLF line endings": {
			give: "package a\r\n\r\ntype T struct {\r\n\t// b\r\n\tb int\r\n\ta int\r\n}\r\n",
			want: "package a\r\n\r\ntype T struct {\r\n\ta int\r\n\t// b\r\n\tb int\r\n}\r\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			have, err := fix.Sort("test.go", []byte(tt.give), fix.Package{})
			must.NoError(t, err)
			must.Eq(t, tt.want, string(have))
			again, err := fix.Sort("test.go", have, fix.Package{})
			must.NoError(t, err, must.Sprint("second run"))
			must.Eq(t, string(have), string(again), must.Sprint("second run changed the result"))
		})
	}
}

func TestSortFieldsTypesInOtherFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, dir, "names.go", "package a\n\ntype Names map[int]string\n")
	give := "package a\n\nvar n = Names{z: \"z\", a: \"a\"}\n\nvar p = Point{y: 1, x: 2}\n"
	writeFile(t, dir, "lookup.go", give)
	have, err := fix.Sort("lookup.go", []byte(give), fix.ScanDir(dir))
	must.NoError(t, err)
	must.Eq(t, "package a\n\nvar n = Names{z: \"z\", a: \"a\"}\n\nvar p = Point{x: 2, y: 1}\n", string(have))
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	must.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}
