Feature: Verify app version

  Scenario Outline: user verify app version success
    Given verify app version request
    When user check app version
    Then returns "OK" status code

  Scenario Outline: user verify app version with invalid request data
    Given verify app version request missing "<missing data>"
    When user check app version
    Then returns "InvalidArgument" status code

    Examples:
      | missing data |
      | packageName  |
      | version      |

  Scenario Outline: user verify app version receive force update error
    Given verify app version request with "<app version>"
    When user check app version
    Then user verify app version receive force update request
    And returns "Aborted" status code

    Examples:
      | app version |
      | 1.-1.0      |
      | 1.-1.1      |