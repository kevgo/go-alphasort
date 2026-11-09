package fix_test

import (
	"testing"

	"github.com/kevgo/go-alphasort/commands/fix"
	"github.com/shoenig/test/must"
)

func TestSortFuncs(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		give string
		want string
	}{
		"already sorted": {
			give: `package a

func alpha() {}

func beta() {}
`,
			want: `package a

func alpha() {}

func beta() {}
`,
		},
		"no functions": {
			give: `package a

var x = 1
`,
			want: `package a

var x = 1
`,
		},
		"single function": {
			give: `package a

func only() {}
`,
			want: `package a

func only() {}
`,
		},
		"reverse order": {
			give: `package a

func charlie() {}

func beta() {}

func alpha() {}
`,
			want: `package a

func alpha() {}

func beta() {}

func charlie() {}
`,
		},
		"doc comments move with their function": {
			give: `package a

// beta does B.
// It has a multi-line comment.
func beta() {}

/*
alpha does A.
*/
func alpha() {}
`,
			want: `package a

/*
alpha does A.
*/
func alpha() {}

// beta does B.
// It has a multi-line comment.
func beta() {}
`,
		},
		"compiler directives move with their function": {
			give: `package a

//go:noinline
func beta() {}

func alpha() {}
`,
			want: `package a

func alpha() {}

//go:noinline
func beta() {}
`,
		},
		"trailing same-line comments move with their function": {
			give: `package a

func beta() {} // about beta

func alpha() {} /* about alpha */
`,
			want: `package a

func alpha() {} /* about alpha */

func beta() {} // about beta
`,
		},
		"floating comments stay in place": {
			give: `package a

// Section: helpers

func beta() {}

// end of helpers

func alpha() {}
`,
			want: `package a

// Section: helpers

func alpha() {}

// end of helpers

func beta() {}
`,
		},
		"methods move behind their type, sorted alphabetically": {
			give: `package a

type T struct{}

func zulu() {}

func (t *T) beta() {}

func (t T) Alpha() {}

func alpha() {}
`,
			want: `package a

type T struct{}

func (t T) Alpha() {}

func (t *T) beta() {}

func alpha() {}

func zulu() {}
`,
		},
		"constructor comes first, then the methods": {
			give: `package a

func helper() {}

// Beta does B.
func (c *Counter) Beta() {}

type Counter struct{ n int }

func (c *Counter) Alpha() {}

// NewCounter creates a Counter.
func NewCounter() *Counter { return &Counter{} }

func aaa() {}
`,
			want: `package a

func aaa() {}

type Counter struct{ n int }

// NewCounter creates a Counter.
func NewCounter() *Counter { return &Counter{} }

func (c *Counter) Alpha() {}

// Beta does B.
func (c *Counter) Beta() {}

func helper() {}
`,
		},
		"constructor without methods": {
			give: `package a

func helper() {}

type T int

func NewT() T { return 0 }
`,
			want: `package a

func helper() {}

type T int

func NewT() T { return 0 }
`,
		},
		"multiple types": {
			give: `package a

type B struct{}

type A struct{}

func (A) M() {}

func NewB() B { return B{} }

func NewA() A { return A{} }

func (B) M() {}
`,
			want: `package a

type B struct{}

func NewB() B { return B{} }

func (B) M() {}

type A struct{}

func NewA() A { return A{} }

func (A) M() {}
`,
		},
		"grouped type declaration": {
			give: `package a

type (
	A struct{}
	B struct{}
)

func (B) M() {}

func NewB() B { return B{} }

func (A) M() {}
`,
			want: `package a

type (
	A struct{}
	B struct{}
)

func (A) M() {}

func NewB() B { return B{} }

func (B) M() {}
`,
		},
		"generic types": {
			give: `package a

type List[T any] struct{ items []T }

func (l *List[T]) Push(item T) {}

func NewList[T any]() *List[T] { return &List[T]{} }

func (l List[T]) Len() int { return len(l.items) }
`,
			want: `package a

type List[T any] struct{ items []T }

func NewList[T any]() *List[T] { return &List[T]{} }

func (l List[T]) Len() int { return len(l.items) }

func (l *List[T]) Push(item T) {}
`,
		},
		"methods and constructors of types declared in other files stay in place": {
			give: `package a

func (o *Other) Zulu() {}

func beta() {}

func (o *Other) Alpha() {}

func NewOther() *Other { return nil }
`,
			want: `package a

func (o *Other) Zulu() {}

func NewOther() *Other { return nil }

func (o *Other) Alpha() {}

func beta() {}
`,
		},
		"only an exact New + type name is a constructor": {
			give: `package a

type T struct{}

func NewTFromString(s string) T { return T{} }

func New() T { return T{} }

func NewT() T { return T{} }
`,
			want: `package a

type T struct{}

func NewT() T { return T{} }

func New() T { return T{} }

func NewTFromString(s string) T { return T{} }
`,
		},
		"methods before the first type declaration": {
			give: `package a

func (T) B() {}

func (T) A() {}

type T struct{}
`,
			want: `package a

type T struct{}

func (T) A() {}

func (T) B() {}
`,
		},
		"floating comments and trailing comments around types": {
			give: `package a

type T struct{} // the T type

// --- helpers ---

func (T) M() {} // method M
`,
			want: `package a

type T struct{} // the T type

func (T) M() {} // method M

// --- helpers ---
`,
		},
		"single-line spacing between methods is preserved": {
			give: `package a

type T struct{}

func (T) A() {}
func (T) B() {}
`,
			want: `package a

type T struct{}

func (T) A() {}
func (T) B() {}
`,
		},
		"CRLF line endings with moved methods": {
			give: "package a\r\n\r\nfunc (T) M() {}\r\n\r\ntype T struct{}\r\n",
			want: "package a\r\n\r\ntype T struct{}\r\n\r\nfunc (T) M() {}\r\n",
		},
		"non-function declarations stay in place": {
			give: `package a

import "fmt"

func beta() { fmt.Println("b") }

const C = 1

func alpha() {}

var V = 2

type T int
`,
			want: `package a

import "fmt"

func alpha() {}

const C = 1

func beta() { fmt.Println("b") }

var V = 2

type T int
`,
		},
		"byte order puts capitalized functions first": {
			give: `package a

func charlie() {}

func Beta() {}

func alpha() {}

func Alpha2() {}
`,
			want: `package a

func Alpha2() {}

func Beta() {}

func alpha() {}

func charlie() {}
`,
		},
		"same name different case uses byte order": {
			give: `package a

func foo() {}

func Foo() {}
`,
			want: `package a

func Foo() {}

func foo() {}
`,
		},
		"multiple init functions keep their relative order": {
			give: `package a

func zulu() {}

func init() { println(1) }

func alpha() {}

func init() { println(2) }
`,
			want: `package a

func alpha() {}

func init() { println(1) }

func init() { println(2) }

func zulu() {}
`,
		},
		"generic functions": {
			give: `package a

func Map[T, U any](in []T, f func(T) U) []U { return nil }

func Filter[T any](in []T, f func(T) bool) []T { return nil }
`,
			want: `package a

func Filter[T any](in []T, f func(T) bool) []T { return nil }

func Map[T, U any](in []T, f func(T) U) []U { return nil }
`,
		},
		"functions of different length": {
			give: `package a

func beta() {
	x := 1
	_ = x
	y := 2
	_ = y
}

func alpha() {}
`,
			want: `package a

func alpha() {}

func beta() {
	x := 1
	_ = x
	y := 2
	_ = y
}
`,
		},
		"comments inside function bodies": {
			give: `package a

func beta() {
	// inside beta
}

func alpha() {
	/* inside alpha */
}
`,
			want: `package a

func alpha() {
	/* inside alpha */
}

func beta() {
	// inside beta
}
`,
		},
		"functions without body": {
			give: `package a

func beta() int

func alpha()
`,
			want: `package a

func alpha()

func beta() int
`,
		},
		"build constraint and package doc stay in place": {
			give: `//go:build linux

// Package a does things.
package a

func beta() {}

func alpha() {}
`,
			want: `//go:build linux

// Package a does things.
package a

func alpha() {}

func beta() {}
`,
		},
		"trailing comment of the last declaration": {
			give: `package a

func beta() {} // b

func alpha() {} // a`,
			want: `package a

func alpha() {} // a

func beta() {} // b`,
		},
		"CRLF line endings": {
			give: "package a\r\n\r\n// b\r\nfunc b() {}\r\n\r\nfunc a() {}\r\n",
			want: "package a\r\n\r\nfunc a() {}\r\n\r\n// b\r\nfunc b() {}\r\n",
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

func TestSortFuncsIdempotent(t *testing.T) {
	t.Parallel()
	give := `package a

// c does C.
func c() {} // trailing c

type T struct{}

func (T) m() {}

func B() {}

var x = 1

func a() {}
`
	once, err := fix.Sort("test.go", []byte(give))
	must.NoError(t, err)
	twice, err := fix.Sort("test.go", once)
	must.NoError(t, err)
	must.Eq(t, string(once), string(twice), must.Sprint("not idempotent"))
}

func TestSortFuncsInvalidSource(t *testing.T) {
	t.Parallel()
	_, err := fix.Sort("test.go", []byte("package a\n\nfunc {"))
	must.Error(t, err)
}
