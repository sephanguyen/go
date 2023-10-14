Feature: Update Assignment

  Background:
    Given <assignment>a signed in "school admin"
    And <assignment>a valid book content
    And user inserts an assignment


  Scenario Outline: authenticate <role> update assignment
    Given <assignment>a signed in "<role>"
    When user updates an assignment
    Then <assignment>returns "<msg>" status code
    Examples:
      | role           | msg              |
      | parent         | PermissionDenied |
      | student        | PermissionDenied |
      | school admin   | OK               |
      | hq staff       | OK               |
      | teacher        | PermissionDenied |
      | centre lead    | PermissionDenied |
      | centre manager | PermissionDenied |
      | teacher lead   | PermissionDenied |

  Scenario: Update Assignment
    And user updates an assignment
    And assignment must be updated
