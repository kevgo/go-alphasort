package fixer_test

import (
	"testing"

	"github.com/kevgo/go-alphasort/fixer"
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
		"methods stay in place": {
			give: `package a

type T struct{}

func zulu() {}

func (t T) Alpha() {}

func (t *T) Beta() {}

func alpha() {}
`,
			want: `package a

type T struct{}

func alpha() {}

func (t T) Alpha() {}

func (t *T) Beta() {}

func zulu() {}
`,
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
			have, err := fixer.SortFuncs("test.go", []byte(tt.give))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(have) != tt.want {
				t.Errorf("mismatch\n--- HAVE ---\n%s\n--- WANT ---\n%s", have, tt.want)
			}
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
	once, err := fixer.SortFuncs("test.go", []byte(give))
	if err != nil {
		t.Fatal(err)
	}
	twice, err := fixer.SortFuncs("test.go", once)
	if err != nil {
		t.Fatal(err)
	}
	if string(once) != string(twice) {
		t.Errorf("not idempotent\n--- FIRST ---\n%s\n--- SECOND ---\n%s", once, twice)
	}
}

func TestSortFuncsInvalidSource(t *testing.T) {
	t.Parallel()
	_, err := fixer.SortFuncs("test.go", []byte("package a\n\nfunc {"))
	if err == nil {
		t.Fatal("expected an error for invalid Go code")
	}
}
