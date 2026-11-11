Feature: sort free-standing functions alphabetically

  Scenario: functions in reverse order
    Given a file "app.go" with content:
      """
      package main

      func main() {}

      func charlie() {}

      func beta() {}

      func alpha() {}
      """
    When I run "alphasort fix app.go"
    Then the exit code is 0
    And the output is empty
    And the file "app.go" now has content:
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
    When I run "alphasort fix lib.go"
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

  Scenario: constructors and methods follow their type
    Given a file "shapes.go" with content:
      """
      package shapes

      import "math"

      func round(f float64) float64 { return math.Round(f) }

      // NewCircleFromDiameter creates a circle with the given diameter.
      func NewCircleFromDiameter(d float64) (*Circle, error) { return &Circle{d / 2}, nil }

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
    Then the exit code is 0
    And the file "shapes.go" now has content:
      """
      package shapes

      import "math"

      func describe(a float64) string { return "" }

      type Circle struct{ R float64 }

      // NewCircle creates a circle.
      func NewCircle(r float64) Circle { return Circle{r} }

      // NewCircleFromDiameter creates a circle with the given diameter.
      func NewCircleFromDiameter(d float64) (*Circle, error) { return &Circle{d / 2}, nil }

      func (c Circle) Area() float64 { return math.Pi * c.R * c.R }

      // Scale enlarges the circle.
      func (c *Circle) Scale(f float64) { c.R *= f }

      const Unit = 1.0

      type Square struct{ S float64 }

      func NewSquare(s float64) Square { return Square{s} }

      func (s Square) Area() float64 { return s.S * s.S }

      func round(f float64) float64 { return math.Round(f) }
      """

  Scenario: functions that don't return an instance of their type first aren't constructors
    Given a file "shapes.go" with content:
      """
      package shapes

      type Circle struct{ R float64 }

      func NewCircleName() string { return "circle" }

      func NewCircles() []Circle { return nil }

      func MakeCircle() Circle { return Circle{} }

      func New() Circle { return Circle{} }
      """
    When I run "alphasort fix shapes.go"
    Then the exit code is 0
    And the file "shapes.go" now has content:
      """
      package shapes

      type Circle struct{ R float64 }

      func MakeCircle() Circle { return Circle{} }

      func New() Circle { return Circle{} }

      func NewCircleName() string { return "circle" }

      func NewCircles() []Circle { return nil }
      """

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
    Then the exit code is 0
    And the file "circle_area.go" now has content:
      """
      package shapes

      func (c Circle) Perimeter() float64 { return 0 }

      func align() {}

      func (c Circle) Area() float64 { return 0 }

      func zoom() {}
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
    When I run "alphasort fix names.go"
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
    When I run "alphasort fix calc.go"
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
    When I run "alphasort fix sorted.go"
    Then the exit code is 0
    And the file "sorted.go" is unchanged

  Scenario: functions in main.go keep their order
    Given a file "main.go" with content:
      """
      package main

      func main() {
      	config := loadConfig()
      	run(config)
      }

      func run(config Config) {}

      func loadConfig() Config { return NewConfig() }

      func (c Config) Valid() bool { return c.Name != "" }

      type Config struct {
      	Name string
      	Debug bool
      }

      func NewConfig() Config { return Config{} }
      """
    When I run "alphasort fix main.go"
    Then the exit code is 0
    And the output is empty
    And the file "main.go" now has content:
      """
      package main

      func main() {
      	config := loadConfig()
      	run(config)
      }

      func run(config Config) {}

      func loadConfig() Config { return NewConfig() }

      type Config struct {
      	Debug bool
      	Name string
      }

      func NewConfig() Config { return Config{} }

      func (c Config) Valid() bool { return c.Name != "" }
      """

  Scenario: main.go in a subdirectory keeps its function order
    Given a file "cmd/tool/main.go" with content:
      """
      package main

      func main() { setup() }

      func setup() {}

      func cleanup() {}
      """
    When I run "alphasort fix"
    Then the exit code is 0
    And the file "cmd/tool/main.go" is unchanged
