@quarantined
Feature: Import location V2

  Background:
    Given some centers
    And some location types

  Scenario: admin try to import valid location
    Given "staff granted role school admin" signin system
    And a valid location payload
    When user import location by csv file
    Then returns "OK" status code
    And the valid location were updated

  Scenario: admin try to import invalid location
    Given "staff granted role school admin" signin system
    And a "<type>" invalid location payload
    When user import location by csv file
    Then returns "InvalidArgument" status code
    And returns error of "<type>" invalid location

    Examples:
      | type                                |
      | no-data                             |
      | wrong-column-count                  |
      | no-name-field                       |
      | no-location_type-field              |
      | no-partner_internal_id-field        |
      | no-partner_internal_parent_id-field |
      | wrong-line-values                   |
      | wrong-partner-values                |
