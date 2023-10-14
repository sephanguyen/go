Feature: Import courses

  Background:
    Given some centers
    And some course types
    And have some courses

  Scenario: admin try to import valid courses
    Given "staff granted role school admin" signin system
    And a valid courses payload
    When user import courses by csv file
    Then returns "OK" status code
    And the valid courses were updated

  Scenario: admin try to import invalid courses
    Given "staff granted role school admin" signin system
    And a "<type>" invalid courses payload
    When user import courses by csv file
    Then returns "InvalidArgument" status code
    And returns error of "<type>" invalid courses

    Examples:
      | type                |
      | no-data             |
      | wrong-column-count  |
      | no-id-field         |
      | no-name-field       |
      | no-type_id-field    |
      | no-remarks-field    |
      | no-partner_id-field |
      | wrong-line-values   |
