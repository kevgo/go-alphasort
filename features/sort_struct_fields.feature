Feature: sort struct fields alphabetically

  Scenario: struct definitions and keyed instantiations
    Given a file "config.go" with content:
      """
      package config

      // Config configures the app.
      type Config struct {
      	// verbose enables more output
      	verbose bool
      	Port    int `json:"port"`
      	Host    string
      	*Logger
      }

      type Logger struct{}

      func Default() Config {
      	return Config{
      		verbose: false,
      		Port:    8080, // the default port
      		Host:    "localhost",
      	}
      }
      """
    When I run "alphasort fix config.go"
    Then the exit code is 0
    And the output is empty
    And the file "config.go" now has content:
      """
      package config

      // Config configures the app.
      type Config struct {
      	Host    string
      	*Logger
      	Port    int `json:"port"`
      	// verbose enables more output
      	verbose bool
      }

      type Logger struct{}

      func Default() Config {
      	return Config{
      		Host:    "localhost",
      		Port:    8080, // the default port
      		verbose: false,
      	}
      }
      """

  Scenario: capitalized fields come first
    Given a file "names.go" with content:
      """
      package names

      type T struct {
      	zulu  int
      	Bravo string
      	alpha bool
      	Alpha bool
      }
      """
    When I run "alphasort fix names.go"
    Then the exit code is 0
    And the output is empty
    And the file "names.go" now has content:
      """
      package names

      type T struct {
      	Alpha bool
      	Bravo string
      	alpha bool
      	zulu  int
      }
      """

  Scenario: comments move with their field, floating comments and blank lines stay in place
    Given a file "server.go" with content:
      """
      package server

      type Server struct {
      	// section: settings

      	// zulu is Z.
      	zulu int // trailing zulu

      	// alpha is A.
      	alpha int /* trailing alpha */
      	// end of settings
      }
      """
    When I run "alphasort fix server.go"
    Then the exit code is 0
    And the output is empty
    And the file "server.go" now has content:
      """
      package server

      type Server struct {
      	// section: settings

      	// alpha is A.
      	alpha int /* trailing alpha */

      	// zulu is Z.
      	zulu int // trailing zulu
      	// end of settings
      }
      """

  Scenario: moved fields keep their spacing
    Given a file "align.go" with content:
      """
      package align

      type T struct {
      	zulu    int
      	charlie string

      	alpha bool
      }
      """
    When I run "alphasort fix align.go"
    Then the exit code is 0
    And the output is empty
    And the file "align.go" now has content:
      """
      package align

      type T struct {
      	alpha bool
      	charlie string

      	zulu    int
      }
      """

  Scenario: names within a multi-name field
    Given a file "multi.go" with content:
      """
      package multi

      type T struct {
      	z    int
      	c, b string
      }
      """
    When I run "alphasort fix multi.go"
    Then the exit code is 0
    And the output is empty
    And the file "multi.go" now has content:
      """
      package multi

      type T struct {
      	b, c string
      	z    int
      }
      """

  Scenario: embedded fields sort by their type name
    Given a file "embed.go" with content:
      """
      package embed

      import "sync"

      type T struct {
      	name string
      	*sync.Mutex
      	Base
      }

      type Base struct{}
      """
    When I run "alphasort fix embed.go"
    Then the exit code is 0
    And the output is empty
    And the file "embed.go" now has content:
      """
      package embed

      import "sync"

      type T struct {
      	Base
      	*sync.Mutex
      	name string
      }

      type Base struct{}
      """

  Scenario: nested and anonymous structs
    Given a file "nested.go" with content:
      """
      package nested

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
      """
    When I run "alphasort fix nested.go"
    Then the exit code is 0
    And the output is empty
    And the file "nested.go" now has content:
      """
      package nested

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
      """

  Scenario: structs with blank fields keep their field order
    Given a file "padding.go" with content:
      """
      package padding

      type T struct {
      	b int32
      	_ [4]byte
      	a int64
      }
      """
    When I run "alphasort fix padding.go"
    Then the exit code is 0
    And the output is empty
    And the file "padding.go" is unchanged

  Scenario: nested and elided instantiations
    Given a file "elided.go" with content:
      """
      package elided

      import "other"

      var x = []Outer{
      	{inner: Inner{d: 1, c: 2}, b: 3},
      	{b: 4},
      }

      var y = map[string]*Inner{"one": {d: 1, c: 2}}

      var z = other.Config{Zulu: 1, Alpha: 2}

      var w = struct{ b, a int }{b: 1, a: 2}
      """
    When I run "alphasort fix elided.go"
    Then the exit code is 0
    And the output is empty
    And the file "elided.go" now has content:
      """
      package elided

      import "other"

      var x = []Outer{
      	{b: 3, inner: Inner{c: 2, d: 1}},
      	{b: 4},
      }

      var y = map[string]*Inner{"one": {c: 2, d: 1}}

      var z = other.Config{Alpha: 2, Zulu: 1}

      var w = struct{ a, b int }{a: 2, b: 1}
      """

  Scenario: map literals with identifier or string keys get sorted
    Given a file "lookup.go" with content:
      """
      package lookup

      type Names map[int]string

      const (
      	z = iota
      	a
      )

      var n = Names{z: "z", a: "a"}

      var s = map[string]int{
      	"zulu":  1, // trailing zulu
      	"Bravo": 2,
      	"alpha": 3,
      }

      var t = []map[string]bool{{"b": true, "a": false}}
      """
    When I run "alphasort fix lookup.go"
    Then the exit code is 0
    And the output is empty
    And the file "lookup.go" now has content:
      """
      package lookup

      type Names map[int]string

      const (
      	z = iota
      	a
      )

      var n = Names{a: "a", z: "z"}

      var s = map[string]int{
      	"Bravo": 2,
      	"alpha": 3,
      	"zulu":  1, // trailing zulu
      }

      var t = []map[string]bool{{"a": false, "b": true}}
      """

  Scenario: array literals, map literals with other keys, and map literals with several side effects keep their order
    Given a file "keep.go" with content:
      """
      package keep

      const (
      	z = iota
      	a
      )

      var s = [...]string{z: "z", a: "a"}

      var m = map[int]string{2: "two", 1: "one"}

      var e = map[string]int{"b": next(), "a": next()}
      """
    When I run "alphasort fix keep.go"
    Then the exit code is 0
    And the output is empty
    And the file "keep.go" is unchanged

  Scenario: instantiations with more than one side effect keep their order
    Given a file "effects.go" with content:
      """
      package effects

      var x = T{b: next(), a: next()}

      var y = T{b: <-ch, a: <-ch}

      var z = T{b: next(), a: len(s), c: int64(1), d: MyInt(2), e: []byte("x"), f: func() int { return next() }}

      type MyInt int
      """
    When I run "alphasort fix effects.go"
    Then the exit code is 0
    And the output is empty
    And the file "effects.go" now has content:
      """
      package effects

      var x = T{b: next(), a: next()}

      var y = T{b: <-ch, a: <-ch}

      var z = T{a: len(s), b: next(), c: int64(1), d: MyInt(2), e: []byte("x"), f: func() int { return next() }}

      type MyInt int
      """

  Scenario: structs instantiated with positional fields in the same file get sorted as well
    Given a file "positional.go" with content:
      """
      package positional

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
      """
    When I run "alphasort fix positional.go"
    Then the exit code is 0
    And the output is empty
    And the file "positional.go" now has content:
      """
      package positional

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
      """

  Scenario: struct fields and functions get sorted together
    Given a file "both.go" with content:
      """
      package both

      func zulu() T { return T{b: 1, a: 2} }

      func alpha() {}

      type T struct {
      	b int
      	a int
      }
      """
    When I run "alphasort fix both.go"
    Then the exit code is 0
    And the output is empty
    And the file "both.go" now has content:
      """
      package both

      func alpha() {}

      func zulu() T { return T{a: 2, b: 1} }

      type T struct {
      	a int
      	b int
      }
      """
