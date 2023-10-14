Feature: Create student
  As a school staff
  I need to be able to create a new student with school histories

  Scenario Outline: Create a student with school histories
    Given generate grade master
    And student info with school histories request and valid "<condition>"
    When "staff granted role school admin" create new student account
    Then new student account created success with school histories
    And receives "OK" status code

    Examples:
      | condition      |
      | one row        |
      | many rows      |
      | mandatory only |

  Scenario Outline: Validate school histories
    Given student info with school histories request and invalid "<condition>"
    When "staff granted role school admin" create new student account
    Then "staff granted role school admin" cannot create that account

    Examples:
      | condition               |
      | missing mandatory       |
      | duplicate school_info   |
      | duplicate school_level  |
      | duplicate school_course |

  Scenario Outline: Create a student with school histories have current school
    Given generate grade master
    And student info with school histories request and valid "<condition>"
    When "staff granted role school admin" create new student account
    Then new student account created success with school histories have current school
    And receives "OK" status code

    Examples:
      | condition                |
      | one row current school   |
#      | many rows current school |
