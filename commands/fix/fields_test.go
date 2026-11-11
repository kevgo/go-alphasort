package fix_test

import (
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
		"array literals keep their order": {
			give: `package a

const (
	z = iota
	a
)

var s = [...]string{z: "z", a: "a"}

var t = []string{z: "z", a: "a"}

type Levels [2]string

var l = Levels{z: "z", a: "a"}
`,
			want: `package a

const (
	z = iota
	a
)

var s = [...]string{z: "z", a: "a"}

var t = []string{z: "z", a: "a"}

type Levels [2]string

var l = Levels{z: "z", a: "a"}
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
		"CRLF line endings": {
			give: "package a\r\n\r\ntype T struct {\r\n\t// b\r\n\tb int\r\n\ta int\r\n}\r\n",
			want: "package a\r\n\r\ntype T struct {\r\n\ta int\r\n\t// b\r\n\tb int\r\n}\r\n",
		},
		"embedded fields get separated from regular fields by a blank line": {
			give: `package a

type T struct {
	Base // trailing Base
	// doc of name
	name string
}

type U struct {
	Base

	name string
}
`,
			want: `package a

type T struct {
	Base // trailing Base

	// doc of name
	name string
}

type U struct {
	Base

	name string
}
`,
		},
		"embedded fields go first, sorted by their type name": {
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
		"embedded fields in CRLF files": {
			give: "package a\r\n\r\ntype T struct {\r\n\tname string\r\n\tBase\r\n}\r\n",
			want: "package a\r\n\r\ntype T struct {\r\n\tBase\r\n\r\n\tname string\r\n}\r\n",
		},
		"embedded fields in single-line structs": {
			give: `package a

type T struct{ name string; Base }
`,
			want: `package a

type T struct{ Base; name string }
`,
		},
		"embedded fields with floating comments in between keep that spacing": {
			give: `package a

type T struct {
	name string
	// end of section

	Base
}
`,
			want: `package a

type T struct {
	Base
	// end of section

	name string
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
		"instantiations with conversions named like parameters of function types": {
			give: `package a

type Meters int

type Callback func(Meters int, len int) (string error)

type Handler interface {
	Handle(Meters int) (len int)
}

func asm(Meters int)

var x = T{b: Meters(1), c: len(s), a: next()}
`,
			want: `package a

type Meters int

type Callback func(Meters int, len int) (string error)

type Handler interface {
	Handle(Meters int) (len int)
}

func asm(Meters int)

var x = T{a: next(), b: Meters(1), c: len(s)}
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
		"instantiations with shadowed or out-of-scope conversions keep their order": {
			give: `package a

type Local int

type Meters int

func f(len func() int) T {
	return T{b: len(), a: next()}
}

func g() T {
	Local := func() int { return next() }
	return T{b: Local(), a: next()}
}

func h() {
	type Remote int
}

func k() T {
	return T{b: Remote(), a: next()}
}

func m() T {
	type Celsius float64
	return T{b: Celsius(1), a: next(), c: Meters(2)}
}
`,
			want: `package a

type Local int

type Meters int

func f(len func() int) T {
	return T{b: len(), a: next()}
}

func g() T {
	Local := func() int { return next() }
	return T{b: Local(), a: next()}
}

func h() {
	type Remote int
}

func k() T {
	return T{b: Remote(), a: next()}
}

func m() T {
	type Celsius float64
	return T{a: next(), b: Celsius(1), c: Meters(2)}
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

var x = T{alpha: 3, Bravo: 2, zulu: 1}

var y = &T{
	// doc of alpha
	alpha: 3,
	zulu: 1, // trailing zulu
}

type T struct {
	alpha int
	Bravo int
	zulu  int
}
`,
		},
		"map literals": {
			give: `package a

import "os"

const (
	z = iota
	a
)

var m = map[int]string{z: "z", a: "a"}

type Names map[int]string

var n = Names{z: "z", a: "a"}

type Tables []map[int]string

var l = Tables{
	{z: "z", a: "a"},
}

var s = map[string]int{
	"zulu":  1, // trailing zulu
	"Bravo": 2,
	"alpha": 3,
}

var r = map[string]int{` + "`b`" + `: 1, "a": 2}

var q = map[any]int{os.Stdout: 1, os.Stderr: 2}
`,
			want: `package a

import "os"

const (
	z = iota
	a
)

var m = map[int]string{a: "a", z: "z"}

type Names map[int]string

var n = Names{a: "a", z: "z"}

type Tables []map[int]string

var l = Tables{
	{a: "a", z: "z"},
}

var s = map[string]int{
	"alpha": 3,
	"Bravo": 2,
	"zulu":  1, // trailing zulu
}

var r = map[string]int{"a": 2, ` + "`b`" + `: 1}

var q = map[any]int{os.Stderr: 2, os.Stdout: 1}
`,
		},
		"map literals with other keys keep their order": {
			give: `package a

var m = map[int]string{2: "two", 1: "one"}

var p = map[[2]int]string{{2, 2}: "b", {1, 1}: "a"}

var c = map[string]int{key(): 1, "a": 2}
`,
			want: `package a

var m = map[int]string{2: "two", 1: "one"}

var p = map[[2]int]string{{2, 2}: "b", {1, 1}: "a"}

var c = map[string]int{key(): 1, "a": 2}
`,
		},
		"map values with several side effects keep their order": {
			give: `package a

var m = map[string]int{"b": next(), "a": next()}
`,
			want: `package a

var m = map[string]int{"b": next(), "a": next()}
`,
		},
		"moved fields keep their spacing": {
			give: `package a

type T struct {
	zulu    int
	charlie string

	alpha bool
}
`,
			want: `package a

type T struct {
	alpha bool
	charlie string

	zulu    int
}
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
	zulu struct {
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
		"single-line struct": {
			give: `package a

type T struct{ b int; a string }
`,
			want: `package a

type T struct{ a string; b int }
`,
		},
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
	alpha bool
	Bravo string
	zulu  int
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
		"tags move with their field": {
			give: "package a\n\ntype T struct {\n\tZ int    `json:\"z\"`\n\tA string `json:\"a\"`\n}\n",
			want: "package a\n\ntype T struct {\n\tA string `json:\"a\"`\n\tZ int    `json:\"z\"`\n}\n",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			have, err := fix.Sort("test.go", []byte(tt.give))
			must.NoError(t, err)
			must.Eq(t, tt.want, string(have))
			again, err := fix.Sort("test.go", have)
			must.NoError(t, err, must.Sprint("second run"))
			must.Eq(t, string(have), string(again), must.Sprint("second run changed the result"))
		})
	}
}
