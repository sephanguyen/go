Feature: Upsert courses

  Background:
    Given a signed in "school admin"
    And there are books existed

  Scenario Outline: Validate upsert courses
    When user creates new "<validity>" courses
    Then returns "<status>" status code
    Examples:
      | validity | status          |
      | valid    | OK              |
      | invalid  | InvalidArgument |

  Scenario: Create courses
    When user creates new "valid" courses
    Then returns "OK" status code

  Scenario: Update courses
    Given there are courses existed
    When user updates "valid" courses
    Then returns "OK" status code
