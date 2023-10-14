Feature: Create student with student phone number and contact preference

  Scenario Outline: Create a student with student phone number
    Given student info with student phone number and contact preference with "<condition>"
    When  "staff granted role school admin" create new student account
    Then new student account created success with student phone number and contact preference
    And receives "OK" status code

    Examples:
      | condition                    |
      | phone and contact preference |
      | contact preference only      |