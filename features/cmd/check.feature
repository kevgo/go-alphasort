Feature: "check" subcommand

  Background:
    Given a file "unsorted.go" with content:
      """
      package demo

      func b() {}

      func a() {}
      """
    And a file "sorted.go" with content:
      """
      package demo

      func a() {}

      func b() {}
      """

  Scenario: list the files that need sorting
    When I run "alphasort check unsorted.go sorted.go"
    Then the exit code is 0
    And the output is:
      """
      unsorted.go
      """
    And the file "unsorted.go" is unchanged
    And the file "sorted.go" is unchanged

  Scenario: all files are sorted
    When I run "alphasort check sorted.go"
    Then the exit code is 0
    And the output is empty

  Scenario: unsorted code via STDIN
    When I run "alphasort check" with input:
      """
      package demo

      func zulu() {}

      func alpha() {}
      """
    Then the exit code is 0
    And the output is:
      """
      <standard input>
      """

  Scenario: sorted code via STDIN
    When I run "alphasort check" with input:
      """
      package demo

      func alpha() {}

      func zulu() {}
      """
    Then the exit code is 0
    And the output is empty

  Scenario: process directories recursively, skipping vendor, testdata, and hidden directories
    Given a file "pkg/sub/deep.go" with content:
      """
      package sub

      func y() {}

      func x() {}
      """
    And a file "pkg/notes.txt" with content:
      """
      func b() {}
      func a() {}
      """
    And a file "vendor/dep/dep.go" with content:
      """
      package dep

      func b() {}

      func a() {}
      """
    And a file "testdata/fixture.go" with content:
      """
      package fixture

      func b() {}

      func a() {}
      """
    And a file ".hidden/hidden.go" with content:
      """
      package hidden

      func b() {}

      func a() {}
      """
    When I run "alphasort check ."
    Then the exit code is 0
    And the output is:
      """
      pkg/sub/deep.go
      unsorted.go
      """
    And the file "pkg/sub/deep.go" is unchanged
    And the file "unsorted.go" is unchanged

  Scenario: invalid Go code
    Given a file "broken.go" with content:
      """
      package broken

      func b() {
      """
    When I run "alphasort check broken.go unsorted.go"
    Then the exit code is 2
    And the error output contains "broken.go"
    And the output is:
      """
      unsorted.go
      """

  Scenario: non-existing file
    When I run "alphasort check zonk.go"
    Then the exit code is 2
    And the error output contains "zonk.go"

  Scenario: unknown flag
    When I run "alphasort check -l unsorted.go"
    Then the exit code is 2
    And the error output contains "unknown flag: -l"
    And the output is empty
