Feature: Import course access paths

  Background:
    Given some centers
    And some course types
    And seeded 20 courses
    And a list of locations are existed in DB
    And have some course access paths

  Scenario: admin imports valid course access paths
    Given "staff granted role school admin" signin system
    And a valid course access paths payload
    When user import course access paths by csv file
    Then returns "OK" status code
    And the valid course access paths were updated

  Scenario Outline: admin imports invalid course access paths
    Given "staff granted role school admin" signin system
    And a "<type>" invalid course access paths payload
    When user import course access paths by csv file
    Then returns "InvalidArgument" status code
    And returns error of "<type>" invalid course access paths

    Examples:
      | type                 |
      | no-data              |
      | wrong-column-count   |
      | no-id-field          |
      | no-course_id-field   |
      | no-location_id-field |
      | wrong-line-values    |
      | not-exist-values     |
