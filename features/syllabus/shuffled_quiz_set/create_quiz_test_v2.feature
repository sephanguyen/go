Feature: Create quiz test

  Background: create quiz background
    Given <shuffled_quiz_set> a signed in "school admin"
    And <shuffled_quiz_set> a valid book content
    And user create a quiz using v2
    And <shuffled_quiz_set> a signed in "student"

  Scenario Outline: auth create a quiz test
    Given <shuffled_quiz_set> a signed in "<role>"
    And school admin add student to a course have a study plan
    When user create quiz test v2
    Then <shuffled_quiz_set> returns "<status code>" status code
    Examples:
      | role           | status code |
      | school admin   | OK          |
      | admin          | OK          |
      | teacher        | OK          |
      | student        | OK          |
      | hq staff       | OK          |
      | center lead    | OK          |
      | center manager | OK          |
      | center staff   | OK          |
      | lead teacher   | OK          |

  Scenario: school admin create a quiz test
    Given <shuffled_quiz_set> a signed in "student"
    And school admin add student to a course have a study plan
    When user create quiz test v2
    Then <shuffled_quiz_set> returns "OK" status code
    And shuffled quiz test have been stored

  Scenario: school admin create a quiz test which have both quiz and question group
    Given <1> existing questions
    Given <2> existing question group
    Given <2> quiz belong to question group
    Given <shuffled_quiz_set> a signed in "student"
    And school admin add student to a course have a study plan
    When user create quiz test v2
    Then <shuffled_quiz_set> returns "OK" status code
    And shuffled quiz test have been stored
    And user got quiz test response
