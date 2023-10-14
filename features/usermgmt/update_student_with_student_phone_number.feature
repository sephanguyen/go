Feature: Update student with student phone number and contact preference
  As a school admin / school staff
  I need to be able to update a existing student

  Scenario Outline: Update a student with student phone number
    Given student info with student phone number and contact preference with "phone and contact preference"
    And "staff granted role school admin" create new student account
    And update student info with student phone number and contact preference with "<condition>"
    When  "staff granted role school admin" update student account
    Then student account updated success with student phone number and contact preference
    And receives "OK" status code

    Examples:
      | condition                    |
      | phone and contact preference |
      | contact preference only      |


  Scenario Outline: Update a student with student phone number with id
    Given student info with student phone number and contact preference with "phone and contact preference"
    And "staff granted role school admin" create new student account
    And update student info with student phone number and contact preference with "<condition>"
    When  "staff granted role school admin" update student account
    Then student account updated success with student phone number id and contact preference
    And receives "OK" status code

    Examples:
      | condition                                   |
      | phone number with id and contact preference |