Feature: Insert Assignment

  Background:
    Given <assignment>a signed in "school admin"
    And <assignment>a valid book content

  Scenario Outline: authenticate <role> insert an assignment
    Given <assignment>a signed in "<role>"
    When user inserts an assignment
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
  #TODO: cover missing field to unit test
  Scenario: Insert Assignment
    When user inserts an assignment
    And assignment must be created
