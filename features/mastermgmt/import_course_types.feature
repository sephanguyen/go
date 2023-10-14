Feature: Import course types

  Background:
    Given some centers
    And some course types

  Scenario: admin try to import valid course types
    Given "staff granted role school admin" signin system
    And a valid course types payload
    When user import course types by csv file
    Then returns "OK" status code
    And the valid course types were updated

  Scenario: admin try to import invalid course types
    Given "staff granted role school admin" signin system
    And a "<type>" invalid course types payload
    When user import course types by csv file
    Then returns "InvalidArgument" status code
    And returns error of "<type>" invalid course types

    Examples:
      | type                 |
      | no-data              |
      | wrong-column-count   |
      | no-id-field          |
      | no-name-field        |
      | no-is_archived-field |
      | no-remarks-field     |
      | wrong-line-values    |