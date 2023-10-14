Feature: Check manabie role in all tables

  Background:
    Given account admin hasura

  Scenario Outline: admin verify role existed
    Given "<service>" name
    When admin export hasura metadata
    Then admin sees the "<role>" existed

    Examples:
    | service        | role    |
    | bob            | MANABIE |
    | eureka         | MANABIE |
    | fatima         | MANABIE |
    | timesheet      | MANABIE |
    | entryexitmgmt  | MANABIE |
    | invoicemgmt    | MANABIE |

  Scenario Outline: admin verify filter and columns for MANABIE role
    Given "<service>" name
    When admin export hasura metadata
    Then admin sees the Manabie role 
    And filters inclued all filters from other roles
    And columns inclued all columns from other roles

    Examples:
    | service        |
    | bob            |
    | eureka         |
    | fatima         |
    | timesheet      |
    | entryexitmgmt  |
    | invoicemgmt    |