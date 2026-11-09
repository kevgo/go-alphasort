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
    When I run "alphasort fix main.go"
    Then the output is empty
    And the file "main.go" now has content:
      """
      package main

      func alpha() {}

      func beta() {}

      func charlie() {}

      func main() {}
      """
    And the exit code is 0

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
    When I run "alphasort fix lib.go"
    Then the file "lib.go" now has content:
      """
      package lib

      // Alpha does A.
      func Alpha() {}

      // Beta does B.
      //
      //go:noinline
      func Beta() {} // trailing comment of Beta
      """
    And the exit code is 0

  Scenario: constructors and methods follow their type
    Given a file "shapes.go" with content:
      """
      package shapes

      import "math"

      func round(f float64) float64 { return math.Round(f) }

      // Scale enlarges the circle.
      func (c *Circle) Scale(f float64) { c.R *= f }

      type Circle struct{ R float64 }

      func NewSquare(s float64) Square { return Square{s} }

      func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

      const Unit = 1.0

      // NewCircle creates a circle.
      func NewCircle(r float64) Circle { return Circle{r} }

      type Square struct{ S float64 }

      func describe(a float64) string { return "" }

      func (s Square) Area() float64 { return s.S * s.S }
      """
    When I run "alphasort fix shapes.go"
    Then the file "shapes.go" now has content:
      """
      package shapes

      import "math"

      func describe(a float64) string { return "" }

      type Circle struct{ R float64 }

      // NewCircle creates a circle.
      func NewCircle(r float64) Circle { return Circle{r} }

      func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

      // Scale enlarges the circle.
      func (c *Circle) Scale(f float64) { c.R *= f }

      const Unit = 1.0

      type Square struct{ S float64 }

      func NewSquare(s float64) Square { return Square{s} }

      func (s Square) Area() float64 { return s.S * s.S }

      func round(f float64) float64 { return math.Round(f) }
      """
    And the exit code is 0

  Scenario: methods of types declared in another file stay in place
    Given a file "circle_area.go" with content:
      """
      package shapes

      func (c Circle) Perimeter() float64 { return 0 }

      func zoom() {}

      func (c Circle) Area() float64 { return 0 }

      func align() {}
      """
    When I run "alphasort fix circle_area.go"
    Then the file "circle_area.go" now has content:
      """
      package shapes

      func (c Circle) Perimeter() float64 { return 0 }

      func align() {}

      func (c Circle) Area() float64 { return 0 }

      func zoom() {}
      """
    And the exit code is 0

  Scenario: capitalized functions come first
    Given a file "names.go" with content:
      """
      package names

      func zeta() {}

      func Gamma() {}

      func beta() {}

      func Alpha() {}
      """
    When I run "alphasort fix names.go"
    Then the file "names.go" now has content:
      """
      package names

      func Alpha() {}

      func Gamma() {}

      func beta() {}

      func zeta() {}
      """
    And the exit code is 0

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
    When I run "alphasort fix calc.go"
    Then the file "calc.go" now has content:
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
    And the exit code is 0

  Scenario: already sorted file
    Given a file "sorted.go" with content:
      """
      package sorted

      func a() {}

      func b() {}
      """
    When I run "alphasort fix sorted.go"
    Then the file "sorted.go" is unchanged
    And the exit code is 0
