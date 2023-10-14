Feature: Create student
  As a school staff
  I need to be able to create a new student with home addresses

  Scenario Outline: Create a student with home addresses
    Given student info with home addresses request and valid "<condition>"
    When "staff granted role school admin" create new student account
    Then new student account created success with home addresses
    And receives "OK" status code

    Examples:
      | condition      |
      | one row        |
      | many rows      |
      | mandatory only |

  Scenario Outline: Validate home addresses
    Given student info with home addresses request and invalid "<condition>"
    When "staff granted role school admin" create new student account
    Then "staff granted role school admin" cannot create that account

    Examples:
      | condition               |
      | incorrect address type  |
