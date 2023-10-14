Feature: Add adhoc assignment

  Scenario Outline: Add adHoc assignment
    Given a valid course background
    And a signed in "student"
    When add adhoc assignment with "<request>"
    Then our system returns "<status code>" status code

    Examples:
      | request            | status code     |
      | missing course_id  | InvalidArgument |
      | missing student_id | InvalidArgument |
      | valid              | OK              |

  Scenario Outline: Add adhoc assignment with times
    Given a valid course background
    And a signed in "student"
    When add adhoc assignment "<times>"
    Then our system must add adhoc assignment

    Examples:
      | times          |
      | one time       |
      | multiple times |

  Scenario: Add adhoc assignment with times
    Given a valid course background
    And a signed in "student"
    When add adhoc assignment with "valid"
    Then our system must add adhoc assignment correctly