Feature: retrieve student comment

  Scenario Outline: a teacher, school admin retrieve comments of a student
    Given a signed in "<signed-in user>"
    And a student with some comments
    When the signed in user retrieve student's comments
    Then get comment belong to user's correctly

  Examples:
      | signed-in user                    |
      | staff granted role school admin   |
      | staff granted role hq staff       |
      | staff granted role centre lead    |
      | staff granted role centre manager |
      | staff granted role centre staff   |
      | staff granted role teacher        |
  
  Scenario Outline: student, parent don't have permission to retrieve comment of a student
    Given a signed in teacher
    And a student with some comments
    And a signed in "<signed-in user>"
    When the signed in user retrieve student's comments
    Then returns "<msg>" status code

    Examples:
      | signed-in user  | msg              |
      | parent          | PermissionDenied |
      | student         | PermissionDenied |
      | unauthenticated | Unauthenticated  |

  Scenario Outline: a teacher retrieve comment of a student with nil studentId(nil studentId)
    Given a signed in teacher
    And a student with some comments
    When the signed in user retrieve student's comment with nil studentId
    Then returns "<msg>" status code

    Examples:
      | msg             |
      | InvalidArgument |
