Feature: Update student
  As a school staff
  I need to be able to create a new student with home addresses

  Scenario Outline: Update student with home addresses
    Given student info with home addresses request and valid "many rows"
    And "staff granted role school admin" create new student account
    And student info with home addresses update request and valid "<condition>"
    When "staff granted role school admin" update student account
    Then student account updated success with home addresses
    And receives "OK" status code

    Examples:
      | condition      |
      | one row        |
      | many rows      |
      | mandatory only |

  Scenario Outline: Validate user addresses
    Given student info with home addresses request and valid "many rows"
    And "staff granted role school admin" create new student account
    And student info with home addresses update request and invalid "<condition>"
    When "staff granted role school admin" update student account
    Then "staff granted role school admin" cannot update student account

    Examples:
      | condition              |
      | incorrect address type |
