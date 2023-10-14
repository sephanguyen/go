Feature: Upsert study plan

  Background:
    Given a signed in "school admin"

  Scenario Outline: Validate upsert study plan
    When user creates new "<validity>" study plan
    Then returns "<status>" status code
    Examples:
      | validity | status          |
      | valid    | OK              |
      | invalid  | InvalidArgument |

  Scenario: Create study plan
    When user creates new "valid" study plan
    Then returns "OK" status code
    And our system must stores correct study plan