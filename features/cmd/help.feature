Feature: "help" subcommand

  Scenario Outline: show the help
    When I run "<invocation>"
    Then the output contains "usage: alphasort <command> [file ...]"
    And the exit code is 0

    Examples:
      | invocation       |
      | alphasort help   |
      | alphasort -h     |
      | alphasort --help |

  Scenario: no subcommand shows the help as an error
    When I run "alphasort"
    Then the output is empty
    And the error output contains "usage: alphasort <command> [file ...]"
    And the exit code is 2

  Scenario: unknown subcommand shows the help as an error
    When I run "alphasort sort main.go"
    Then the output is empty
    And the error output contains "unknown command: sort"
    And the error output contains "usage: alphasort <command> [file ...]"
    And the exit code is 2
