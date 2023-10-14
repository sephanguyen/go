@major
Feature: Retrieve Student Entry and Exit records
  As a parent
  I am able to see entry exit records of my children on learner app
  Background:
    Given "parent" logins Learner App
    And this parent has an existing student

  Scenario Outline: Parent selects student with entry and exit records
    Given student has "existing" entry and exit "<date-record>" record
    When parent is at the entry exit records screen
    And parent selects this existing student
    And parent checks the filter for records "<date-record>"
    Then records found with default limit are displayed successfully
    And receives "OK" status code

    Examples:
      | date-record |
      | This month  |
      | This year   |
      | Last month  |

  Scenario: Parent checks all the entry and exit records of the student
    Given student has "existing" entry and exit "<date-record>" record
    And parent is at the entry exit records screen
    And parent selects this existing student
    And records found with default limit are displayed successfully
    When parent scrolls down to display all records
    Then all records found are displayed successfully
    And receives "OK" status code

    Examples:
      | date-record |
      | This month  |
      | This year   |
      | Last month  |

  Scenario Outline: Parent selects student with no entry and exit records
    Given student has "no entry and exit" record
    When parent is at the entry exit records screen
    And parent selects this existing student
    Then no records found displayed successfully
    And receives "OK" status code

  Scenario Outline: Parent selects another student and entry and exit records displayed successfully
    Given parent has another existing student
    And this student has "existing" entry and exit record "<date-record>"
    And parent is at the entry exit records screen
    When parent selects this existing student
    Then records found with default limit are displayed successfully
    And receives "OK" status code

    Examples:
      | date-record |
      | This month  |
      | This year   |
      | Last month  |
      | None        |

  Scenario Outline: Parent selects student with entry and exit records last year
    Given student has "existing" entry and exit "<date-record>" record
    When parent is at the entry exit records screen
    And parent selects this existing student
    And parent checks the filter for records "This year"
    Then no records found displayed successfully
    And receives "OK" status code

    Examples:
      | date-record |
      | Last year   |

  Scenario Outline: Parent selects student with entry and exit records last year this month
    Given student has "existing" entry and exit "<date-record>" record
    When parent is at the entry exit records screen
    And parent selects this existing student
    And parent checks the filter for records "This month"
    Then no records found displayed successfully
    And receives "OK" status code

    Examples:
      | date-record          |
      | Last year this month |