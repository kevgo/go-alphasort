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
      	Host string
      	*Logger
      	Port int `json:"port"`
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

  Scenario: structs instantiated with positional fields in another file of the package keep their field order
    Given a file "point.go" with content:
      """
      package geo

      type Point struct {
      	y int
      	x int
      }
      """
    And a file "point_test.go" with content:
      """
      package geo

      var origin = Point{0, 0}
      """
    When I run "alphasort fix point.go"
    Then the exit code is 0
    And the file "point.go" is unchanged
