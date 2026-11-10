Feature: "version" subcommand

  Scenario Outline: show the version
    When I run "<invocation>"
    Then the exit code is 0
    And the output matches "^alphasort \d+\.\d+\.\d+\S*\n$"
    And the error output is empty

    Examples:
      | invocation          |
      | alphasort version   |
      | alphasort -version  |
      | alphasort --version |

  Scenario: unexpected argument
    When I run "alphasort version zonk"
    Then the exit code is 2
    And the output is empty
    And the error output contains "unexpected argument: zonk"
