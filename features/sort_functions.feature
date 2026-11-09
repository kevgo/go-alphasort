Feature: sort free-standing functions alphabetically

  Scenario: functions in reverse order
    Given a file "main.go" with content:
      """
      package main

      func main() {}

      func charlie() {}

      func beta() {}

      func alpha() {}
      """
    When I run "go-alphasort -w main.go"
    Then the exit code is 0
    And the output is empty
    And the file "main.go" now has content:
      """
      package main

      func alpha() {}

      func beta() {}

      func charlie() {}

      func main() {}
      """

  Scenario: doc comments, directives, and trailing comments move with their function
    Given a file "lib.go" with content:
      """
      package lib

      // Beta does B.
      //
      //go:noinline
      func Beta() {} // trailing comment of Beta

      // Alpha does A.
      func Alpha() {}
      """
    When I run "go-alphasort -w lib.go"
    Then the exit code is 0
    And the file "lib.go" now has content:
      """
      package lib

      // Alpha does A.
      func Alpha() {}

      // Beta does B.
      //
      //go:noinline
      func Beta() {} // trailing comment of Beta
      """

  Scenario: methods and other declarations keep their position
    Given a file "shapes.go" with content:
      """
      package shapes

      import "math"

      type Circle struct{ R float64 }

      func NewSquare(s float64) Square { return Square{s} }

      func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

      const Unit = 1.0

      func NewCircle(r float64) Circle { return Circle{r} }

      type Square struct{ S float64 }

      func (s Square) Area() float64 { return s.S * s.S }
      """
    When I run "go-alphasort -w shapes.go"
    Then the exit code is 0
    And the file "shapes.go" now has content:
      """
      package shapes

      import "math"

      type Circle struct{ R float64 }

      func NewCircle(r float64) Circle { return Circle{r} }

      func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

      const Unit = 1.0

      func NewSquare(s float64) Square { return Square{s} }

      type Square struct{ S float64 }

      func (s Square) Area() float64 { return s.S * s.S }
      """

  Scenario: capitalized functions come first
    Given a file "names.go" with content:
      """
      package names

      func zeta() {}

      func Gamma() {}

      func beta() {}

      func Alpha() {}
      """
    When I run "go-alphasort -w names.go"
    Then the exit code is 0
    And the file "names.go" now has content:
      """
      package names

      func Alpha() {}

      func Gamma() {}

      func beta() {}

      func zeta() {}
      """

  Scenario: multi-line function bodies
    Given a file "calc.go" with content:
      """
      package calc

      func Sum(values ...int) int {
      	total := 0
      	for _, value := range values {
      		total += value
      	}
      	return total
      }

      func Double(value int) int {
      	return value * 2
      }
      """
    When I run "go-alphasort -w calc.go"
    Then the exit code is 0
    And the file "calc.go" now has content:
      """
      package calc

      func Double(value int) int {
      	return value * 2
      }

      func Sum(values ...int) int {
      	total := 0
      	for _, value := range values {
      		total += value
      	}
      	return total
      }
      """

  Scenario: already sorted file
    Given a file "sorted.go" with content:
      """
      package sorted

      func a() {}

      func b() {}
      """
    When I run "go-alphasort -w sorted.go"
    Then the exit code is 0
    And the file "sorted.go" is unchanged
