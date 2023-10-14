@quarantined
Feature: Import location type V2

  Background:
    Given some centers
    And some location types

  Scenario: admin try to import valid location type
    Given "staff granted role school admin" signin system
    And a valid location type payload
    When user import location type by csv file
    Then returns "OK" status code
    And the valid location type were updated

  Scenario: admin try to import invalid location type
    Given "staff granted role school admin" signin system
    And a "<type>" invalid location type payload
    When user import location type by csv file
    Then returns "InvalidArgument" status code
    And returns error of "<type>" invalid location type

    Examples:
      | type                  |
      | no-data               |
      | wrong-column-count    |
      | no-name-field         |
      | no-display_name-field |
      | no-level-field        |
      | wrong-line-values     |
      | level-already-existed |
      | swapped-level         |
