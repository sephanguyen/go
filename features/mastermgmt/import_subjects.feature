Feature: Import subjects

  Scenario: admin imports valid subjects
    Given "staff granted role school admin" signin system
    And a valid subjects payload
    When user import subjects by csv file
    Then returns "OK" status code
    And the valid subjects were updated

  Scenario Outline: admin imports invalid subjects
    Given "staff granted role school admin" signin system
    And a "<type>" invalid subjects payload
    When user import subjects by csv file
    Then returns "InvalidArgument" status code
    And returns error of "<type>" invalid subjects

    Examples:
      | type               |
      | no-data            |
      | wrong-column-count |
      | no-id-field        |
      | no-name-field      |
      | wrong-line-values  |
