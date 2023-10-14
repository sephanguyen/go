Feature: Create quiz test

  Background: create quiz background
    Given a signed in "school admin"
    And a valid book content
    And user create a quiz using v2
    Then returns "OK" status code

  Scenario Outline: Authenticate create quiz test
    And a signed in "<role>"
    When user create quiz test
    Then returns "<status code>" status code
    Examples:
      | role           | status code      |
      | parent         | PermissionDenied |
      | student        | OK               |
      | teacher        | PermissionDenied |
      | center lead    | PermissionDenied |
      | center manager | PermissionDenied |
      | center staff   | PermissionDenied |
      | hq staff       | PermissionDenied |
      | school admin   | PermissionDenied |

  Scenario: student create a quiz test
    Given a signed in "student"
    And school admin add student to a course have a study plan
    When student create quiz test
    Then returns "OK" status code
    And our system must returns quizzes correctly

  Scenario: student create a quiz test which have both quiz and question group
    Given <2> existing question group
    And <2> quiz belong to question group
    And a signed in "student"
    And school admin add student to a course have a study plan
    When student create quiz test
    Then returns "OK" status code
    And user got quiz test response