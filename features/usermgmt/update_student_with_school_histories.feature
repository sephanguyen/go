Feature: Update student with school histories
  As a school admin / school staff
  I need to be able to update a existing student

  Scenario Outline: Update student with school histories
    Given generate grade master
    And student info with school histories request and valid "many rows"
    And "staff granted role school admin" create new student account
    And student info with school histories update request and valid "<condition>"
    When "staff granted role school admin" update student account
    Then student account updated success with school histories
    And receives "OK" status code

    Examples:
      | condition      |
      | one row        |
      | many rows      |
      | mandatory only |

  Scenario Outline: Validate school histories
    Given generate grade master
    And student info with school histories request and valid "many rows"
    And "staff granted role school admin" create new student account
    And student info with school histories update request and invalid "<condition>"
    When "staff granted role school admin" update student account
    Then "staff granted role school admin" cannot update student account

    Examples:
      | condition               |
      | missing mandatory       |
      | duplicate school_info   |
      | duplicate school_level  |
      | duplicate school_course |

  Scenario Outline: Update student with school histories
    Given generate grade master
    And student info with school histories request and valid "many rows"
    And "staff granted role school admin" create new student account
    And student info with school histories update request and valid "<condition>"
    When "staff granted role school admin" update student account
    Then student account updated success with school histories have current school
    And receives "OK" status code

    Examples:
      | condition              |
      | one row current school |

  Scenario Outline: Update student with school histories
    Given generate grade master
    And student info with school histories request and valid "one row current school"
    And "staff granted role school admin" create new student account
    And student info with school histories update request and valid "one row"
    When "staff granted role school admin" update student account
    Then student account updated success with school histories remove current school
    And receives "OK" status code

    Examples:
      | condition              |
      | one row current school |