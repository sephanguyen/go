Feature: Import grades

  Scenario: admin imports valid grades
    Given "staff granted role school admin" signin system
    And a valid grades payload
    When user import grades by csv file
    Then returns "OK" status code
    And the valid grades were updated

  Scenario: admin imports invalid grades
    Given "staff granted role school admin" signin system
    And a "<type>" invalid grades payload
    When user import grades by csv file
    Then returns "InvalidArgument" status code
    And returns error of "<type>" invalid grades

    Examples:
      | type                 |
      | no-data              |
      | wrong-column-count   |
      | no-id-field          |
      | no-name-field        |
      | no-partner_id-field  |
      | no-sequence-field    |
      | no-remarks-field     |
      | wrong-line-values    |
