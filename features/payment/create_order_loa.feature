Feature: LOA request
  @quarantined
  Scenario Outline: LOA request success
    Given prepare data for LOA request with "<valid data>"
    When "school admin" submit order
    Then LOA request is created successfully
    And receives "OK" status code

    Examples:
      | valid data                                              |
      | valid LOA request with disabled prorating               |
      | valid LOA request with prorating and discount           |
      | empty billed at order with prorating and discount       |
      | empty upcoming billing with prorating and discount      |
      | no active recurring products                            |

  @quarantined
  Scenario Outline: LOA request with pausable tag
    Given prepare data for LOA request with product pausable tag set to "<value>"
    When "school admin" submit order
    Then product setting pausable tag "<value>" validated successfully

    Examples:
      | value              |
      | true               |
      | false              |