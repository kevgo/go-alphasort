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

  Scenario: print the sorted file to STDOUT without modifying it
    When I run "go-alphasort unsorted.go"
    Then the exit code is 0
    And the output is:
      """
      package demo
      
      func a() {}
      
      func b() {}
      """
    And the file "unsorted.go" is unchanged

  Scenario: read from STDIN and write to STDOUT
    When I run "go-alphasort" with input:
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

  Scenario: list files that need sorting
    When I run "go-alphasort -l unsorted.go sorted.go"
    Then the exit code is 0
    And the output is:
      """
      unsorted.go
      """
    And the file "unsorted.go" is unchanged

  Scenario: list and fix files at the same time
    When I run "go-alphasort -l -w unsorted.go sorted.go"
    Then the exit code is 0
    And the output is:
      """
      unsorted.go
      """
    And the file "unsorted.go" now has content:
      """
      package demo
      
      func a() {}
      
      func b() {}
      """
    And the file "sorted.go" is unchanged

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
    When I run "go-alphasort -l -w ."
    Then the exit code is 0
    And the output is:
      """
      pkg/sub/deep.go
      unsorted.go
      """
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
    When I run "go-alphasort -w broken.go unsorted.go"
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
    When I run "go-alphasort zonk.go"
    Then the exit code is 2
    And the error output contains "zonk.go"

  Scenario: writing is not possible when reading from STDIN
    When I run "go-alphasort -w" with input:
      """
      package demo
      """
    Then the exit code is 2
    And the error output contains "cannot use -w with standard input"

  Scenario: help
    When I run "go-alphasort -h"
    Then the exit code is 0
    And the error output contains "usage: go-alphasort [flags] [path ...]"
