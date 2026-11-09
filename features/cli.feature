Feature: command-line interface

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

  Scenario: check lists the files that need sorting
    When I run "alphasort check unsorted.go sorted.go"
    Then the exit code is 0
    And the output is:
      """
      unsorted.go
      """
    And the file "unsorted.go" is unchanged
    And the file "sorted.go" is unchanged

  Scenario: fix sorts the files in place
    When I run "alphasort fix unsorted.go sorted.go"
    Then the exit code is 0
    And the output is empty
    And the file "unsorted.go" now has content:
      """
      package demo

      func a() {}

      func b() {}
      """
    And the file "sorted.go" is unchanged

  Scenario: check reads from STDIN
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

  Scenario: fix reads from STDIN and writes to STDOUT
    When I run "alphasort fix" with input:
      """
      package demo

      func zulu() {}

      func alpha() {}
      """
    Then the exit code is 0
    And the output is:
      """
      package demo

      func alpha() {}

      func zulu() {}
      """

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
    When I run "alphasort fix ."
    Then the exit code is 0
    And the file "pkg/sub/deep.go" now has content:
      """
      package sub

      func x() {}

      func y() {}
      """
    And the file "pkg/notes.txt" is unchanged
    And the file "vendor/dep/dep.go" is unchanged
    And the file "testdata/fixture.go" is unchanged
    And the file ".hidden/hidden.go" is unchanged

  Scenario: invalid Go code
    Given a file "broken.go" with content:
      """
      package broken

      func b() {
      """
    When I run "alphasort fix broken.go unsorted.go"
    Then the exit code is 2
    And the error output contains "broken.go"
    And the file "broken.go" is unchanged
    And the file "unsorted.go" now has content:
      """
      package demo

      func a() {}

      func b() {}
      """

  Scenario: non-existing file
    When I run "alphasort check zonk.go"
    Then the exit code is 2
    And the error output contains "zonk.go"

  Scenario Outline: help
    When I run "<invocation>"
    Then the exit code is 0
    And the output contains "usage: alphasort <command> [path ...]"

    Examples:
      | invocation       |
      | alphasort help   |
      | alphasort -h     |
      | alphasort --help |

  Scenario: no command
    When I run "alphasort"
    Then the exit code is 2
    And the error output contains "usage: alphasort <command> [path ...]"

  Scenario: unknown command
    When I run "alphasort sort unsorted.go"
    Then the exit code is 2
    And the error output contains "unknown command: sort"
    And the file "unsorted.go" is unchanged

  Scenario: unknown flag
    When I run "alphasort fix -w unsorted.go"
    Then the exit code is 2
    And the error output contains "unknown flag: -w"
    And the file "unsorted.go" is unchanged
