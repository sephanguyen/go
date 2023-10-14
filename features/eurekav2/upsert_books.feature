@blocker
Feature: Upsert books

  Background:
    Given a signed in "school admin"

  Scenario Outline: Validate upsert books
    When user creates new "<validity>" books
    Then returns "<status>" status code
    Examples:
      | validity | status          |
      | valid    | OK              |
      | invalid  | InvalidArgument |

  Scenario: Create books
    When user creates new "valid" books
    Then returns "OK" status code
    And our system must stores correct books

  Scenario: Update books
    Given there are books existed
    When user updates "valid" books
    Then returns "OK" status code
    And our system must update the books correctly
