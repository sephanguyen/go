Feature: Remove parents
  As a school staff
  I'm able to remove parents
  Background: sign in with role staff granted role school admin
    Given a signed in "staff granted role school admin"

  Scenario Outline: delete parents successfully
    Given create data parent with 2 different students
    When "staff granted role <role>" remove new parents with "valid condition"
    Then parents were removed successfully

    Examples:
      | role           |
      | school admin   |
      | hq staff       |
      | centre lead    |
      | centre manager |
      | centre staff   |

  Scenario Outline: teacher, student, parent don't have permission to remove parent from student
    Given parents data to remove from student
    When "<signed-in user>" remove new parents with "valid condition"
    Then returns "PermissionDenied" status code

    Examples:
      | signed-in user |
      | teacher        |
      | student        |
      | parent         |

  Scenario Outline: cannot remove parent if parent id is empty
    Given parents data to remove from student
    When "staff granted role school admin" remove new parents with "invalid <condition>"
    Then receives "InvalidArgument" status code

    Examples:
      | condition          |
      | empty parentID     |
      | empty studentID    |
      | un-exist parentID  |
      | un-exist studentID |

  Scenario: cannot remove parent if student don't have relationship
    Given parents data without relationship to remove
    When "staff granted role school admin" remove new parents with "valid condition"
    Then receives "InvalidArgument" status code

  Scenario: cannot remove parent when parent has only student with the enable feature flag
    Given parents data to remove from student
    When "staff granted role school admin" remove new parents with "valid condition"
    Then receives "InvalidArgument" status code and message "invalidRemoveParent"
    And parents was not removed in database

# Scenario: can remove parent when parent has only student with the disable feature flag
#   Given "disable" Unleash feature with feature name "User_StudentManagement_BackOffice_ParentLocationValidation"
#   And parents data to remove from student
#   When "staff granted role school admin" remove new parents with "valid condition"
#   Then parents were removed successfully
