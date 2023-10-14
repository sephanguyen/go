@quarantined
Feature: Migrate Invoice Bill Item

  Scenario Outline: Job script migrate invoice bill item
    Given admin is logged-in back office on organization "<organization-id>"
    And there are "3" existing invoice and bill_item that have the same invoice reference
    When an admin runs the migrate invoice bill item job script
    Then the migrate invoice bill item script has no error
    And the invoice bill items were successfully migrated
    And the migrated invoice bill item have the same reference

    Examples:
      | organization-id |
      | -2147483642     |
      | -2147483635     |

  Scenario Outline: Job script migrate invoice bill item with invalid parameters
    When an admin runs the migrate invoice bill item job script with "<condition>"
    Then the migrate invoice bill item script returns error

    Examples:
      | condition     |
      | empty-orgID   |
      | empty-userID  |
      | invalid-orgID |
