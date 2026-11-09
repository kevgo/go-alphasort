Feature: "fix" subcommand

  Background:
    Given a file "sorted.go" with content:
      """
      package demo

      func a() {}

      func b() {}
      """
    And a file "unsorted.go" with content:
      """
      package demo

      func b() {}

      func a() {}
      """

  Scenario: fix only the given files
    Given a file "other.go" with content:
      """
      package demo

      func d() {}

      func c() {}
      """
    When I run "alphasort fix unsorted.go sorted.go"
    Then the output is empty
    And the file "unsorted.go" now has content:
      """
      package demo

      func a() {}

      func b() {}
      """
    And the file "other.go" is unchanged
    And the file "sorted.go" is unchanged
    And the exit code is 0

  Scenario: without arguments, fix all Go files in the current directory and its subdirectories
    Given a file ".hidden/hidden.go" with content:
      """
      package hidden

      func b() {}

      func a() {}
      """
    And a file "_build/build.go" with content:
      """
      package build

      func b() {}

      func a() {}
      """
    And a file "pkg/notes.txt" with content:
      """
      func b() {}
      func a() {}
      """
    And a file "pkg/sub/deep.go" with content:
      """
      package sub

      func y() {}

      func x() {}
      """
    And a file "testdata/fixture.go" with content:
      """
      package fixture

      func b() {}

      func a() {}
      """
    And a file "vendor/dep/dep.go" with content:
      """
      package dep

      func b() {}

      func a() {}
      """
    When I run "alphasort fix"
    Then the output is empty
    And the file "pkg/sub/deep.go" now has content:
      """
      package sub

      func x() {}

      func y() {}
      """
    And the file "unsorted.go" now has content:
      """
      package demo

      func a() {}

      func b() {}
      """
    And the file ".hidden/hidden.go" is unchanged
    And the file "_build/build.go" is unchanged
    And the file "pkg/notes.txt" is unchanged
    And the file "sorted.go" is unchanged
    And the file "testdata/fixture.go" is unchanged
    And the file "vendor/dep/dep.go" is unchanged
    And the exit code is 0

  Scenario: explicitly given files get fixed even in skipped directories
    Given a file "vendor/dep/dep.go" with content:
      """
      package dep

      func b() {}

      func a() {}
      """
    When I run "alphasort fix vendor/dep/dep.go"
    Then the file "vendor/dep/dep.go" now has content:
      """
      package dep

      func a() {}

      func b() {}
      """
    And the file "unsorted.go" is unchanged
    And the exit code is 0

  Scenario: directories are not accepted as arguments
    Given a file "pkg/lib.go" with content:
      """
      package pkg

      func b() {}

      func a() {}
      """
    When I run "alphasort fix pkg"
    Then the error output contains "pkg is a directory: please provide files, or no arguments to fix all files"
    And the file "pkg/lib.go" is unchanged
    And the exit code is 2

  Scenario: invalid Go code in a given file
    Given a file "broken.go" with content:
      """
      package broken

      func b() {
      """
    When I run "alphasort fix broken.go unsorted.go"
    Then the error output contains "broken.go"
    And the file "unsorted.go" now has content:
      """
      package demo

      func a() {}

      func b() {}
      """
    And the file "broken.go" is unchanged
    And the exit code is 2

  Scenario: invalid Go code without arguments does not stop the other files from getting fixed
    Given a file "a_broken.go" with content:
      """
      package demo

      func b() {
      """
    When I run "alphasort fix"
    Then the error output contains "a_broken.go"
    And the file "unsorted.go" now has content:
      """
      package demo

      func a() {}

      func b() {}
      """
    And the file "a_broken.go" is unchanged
    And the exit code is 2

  Scenario: non-existing file
    When I run "alphasort fix zonk.go"
    Then the error output contains "zonk.go"
    And the exit code is 2

  Scenario: unknown flag
    When I run "alphasort fix -w unsorted.go"
    Then the error output contains "unknown flag: -w"
    And the file "unsorted.go" is unchanged
    And the exit code is 2
