Feature: Add books to a course
  Background:
    Given a signed in "school admin"
    And there are books existed
    And a valid course background

  Scenario Outline: Authentication for validate add books
    Given a signed in "<role>"
    And an "valid" add books request
    When user try to add books to course
    Then returns "<status>" status code

    Examples: 
      | role        |status          |
      |school admin |OK              |
      |student      |PermissionDenied|
      |parent       |PermissionDenied|
      |teacher      |OK              |
      |hq staff     |OK              |
  
  Scenario Outline: Validate add books
        Given an "<validity>" add books request
        When user try to add books to course
        Then returns "<status>" status code

        Examples:
            | validity            | status          |
            | valid               | OK              |
            | non-existed bookIDs | NotFound        |
            | empty bookIDs       | InvalidArgument |
            | empty courseID      | InvalidArgument |

  Scenario: User add valid books to a course
    Given an "valid" add books request
    When user try to add books to course
    And our system must adds books to course correctly
