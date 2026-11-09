Feature: "help" subcommand

  Scenario Outline: show the help
    When I run "<invocation>"
    Then the exit code is 0
    And the output contains "usage: alphasort <command> [path ...]"

    Examples:
      | invocation       |
      | alphasort help   |
      | alphasort -h     |
      | alphasort --help |

  Scenario: no subcommand shows the help as an error
    When I run "alphasort"
    Then the exit code is 2
    And the output is empty
    And the error output contains "usage: alphasort <command> [path ...]"

  Scenario: unknown subcommand shows the help as an error
    When I run "alphasort sort main.go"
    Then the exit code is 2
    And the output is empty
    And the error output contains "unknown command: sort"
    And the error output contains "usage: alphasort <command> [path ...]"
