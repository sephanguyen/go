@blocker
Feature: Open API Upsert Student With Enrollment Status History

  Background: Sign in with role "staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario Outline: Create students by OpenAPI with "<condition>" unsuccessfully
    When school admin creates 1 students with "<condition>" by OpenAPI in folder "enrollment_status_histories"
    Then student were created unsuccessfully by OpenAPI with code "40004" and field "<field>"

    Examples:
      | condition                                                 | field             |
      | enrollment status temporary only                          | enrollment_status |
      | 1 enrollment status but empty location                    | location          |
      | 2 enrollment status but 1 location                        | location          |
      | enrollment status temporary with start date in future     | start_date        |
      | enrollment status potential with start date in future     | start_date        |
      | enrollment status non-potential with start date in future | start_date        |

  Scenario Outline: Update students by OpenAPI with "<condition>" to "changing to enrollment status non-potential" unsuccessfully
    Given school admin creates 1 students with "<condition>" by OpenAPI in folder "enrollment_status_histories"
    When school admin updates 1 students with "<update_condition>" by OpenAPI
    Then student were updated unsuccessfully by OpenAPI with code "40004" and field "<field>"

    Examples:
      | condition                   | update_condition                                                           | field      |
      | enrollment status potential | adding new location with enrollment status potential with feature date     | start_date |
      | enrollment status potential | adding new location with enrollment status temporary with feature date     | start_date |
      | enrollment status potential | adding new location with enrollment status non-potential with feature date | start_date |

  Scenario Outline: Update students by OpenAPI csv file with "<condition>" to "<desired-enrollment_status>" successfully
    Given school admin creates 1 students with "<condition>" by OpenAPI in folder "enrollment_status_histories"
    When school admin updates 1 students with "<desired-enrollment_status>" by OpenAPI
    Then students were upserted successfully by OpenAPI

    Examples:
      | condition                   | desired-enrollment_status                            |
      | enrollment status potential | changing to enrollment status graduate and new date  |
      | enrollment status potential | changing to enrollment status withdraw and new date  |
      | enrollment status potential | changing to enrollment status LOA and new date       |
      | enrollment status potential | changing to enrollment status non-potential new date |
      | enrollment status potential | changing to enrollment status enrolled new date      |
      | enrollment status enrolled  | changing to enrollment status potential and new date |
      | enrollment status enrolled  | changing to enrollment status withdraw and new date  |
      | enrollment status enrolled  | changing to enrollment status graduate and new date  |
      | enrollment status enrolled  | changing to enrollment status LOA and new date       |
      | enrollment status temporary | changing to enrollment status potential and new date |
      | enrollment status enrolled  | changing to enrollment status non-potential new date |
