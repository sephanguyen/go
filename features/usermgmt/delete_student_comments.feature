Feature: delete comment in student

  The teacher delete student's comment

  Scenario: a teacher delete comment of a student
    Given a signed in "<signed-in user>"
    And a student with some comments
    When the signed in user delete student's comment
    And our systems have to store comment correctly

    Examples:
      | signed-in user                    |
      | staff granted role school admin   |
      | staff granted role hq staff       |
      | staff granted role centre lead    |
      | staff granted role centre manager |
      | staff granted role centre staff   |
      | staff granted role teacher        |

  Scenario Outline: student, parent don't have permission to delete comment of a student
      Given a signed in teacher
      And a student with some comments
      And a signed in "<signed-in user>"
      When the signed in user delete student's comment
      Then returns "<msg>" status code

      Examples:
        | signed-in user  | msg              |
        | parent          | PermissionDenied |
        | student         | PermissionDenied |
        | unauthenticated | Unauthenticated  |

  Scenario Outline: a teacher delete comment of a student with invalid parameter(nil commentIds)
    Given a signed in teacher
    And a student with some comments
    When the signed in user delete student's comment with nil commentIds
    Then returns "<msg>" status code

    Examples:
      | msg             |
      | InvalidArgument |

  Scenario Outline: a teacher delete comment of a student but comment not exist
    Given a signed in teacher
    And a student with some comments
    When the signed in user delete student's comment but comment not exist
    Then returns "<msg>" status code

    Examples:
      | msg |
      | OK  |
