Feature: Upsert chapters

  Background:
    Given a signed in "school admin"
    And user has created an empty book

  Scenario: Validate upsert chapters
    When user creates new "<validity>" chapters
    Then returns "<status>" status code
    Examples:
      | validity | status          |
      | valid    | OK              |
      | invalid  | InvalidArgument |

  Scenario: Create chapters
    When user creates new "valid" chapters
    Then returns "OK" status code
    And our system must stores correct chapters

  Scenario: Update chapters
    Given there are chapters existed
    When user updates "valid" chapters
    Then returns "OK" status code
    And our system must update the chapters correctly
