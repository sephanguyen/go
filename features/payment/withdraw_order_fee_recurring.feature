Feature: Withdraw order Recurring Fee

  Scenario Outline: Withdraw order recurring fee success
    Given prepare data for withdraw order recurring fee with "<valid data>"
    When "school admin" submit order
    Then withdraw order recurring fee is created "<status>" for "<valid data>"
    And receives "<status code>" status code

    Examples:
      | valid data                                         | status         | status code        |
      | valid withdrawal request with disabled prorating   | successfully   | OK                 |
      | valid graduate request with disabled prorating     | successfully   | OK                 |
      | empty billed at order disabled prorating           | successfully   | OK                 |
      | empty upcoming billing disabled prorating          | successfully   | OK                 |
      | empty billed at order with prorating and discount  | successfully   | OK                 |
      | empty upcoming billing with prorating and discount | successfully   | OK                 |
      | empty billing items                                | successfully   | OK                 |
      | out of version                                     | unsuccessfully | FailedPrecondition |
      | non-enrolled status                                | unsuccessfully | FailedPrecondition |