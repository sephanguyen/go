@blocker
Feature: Import Student Enrollment Status History

    Background: Sign in with role "staff granted role school admin"
        Given a signed in "staff granted role school admin"

    Scenario Outline: Create students by import csv file with "<condition>" successfully
        When school admin create 1 students with "<condition>" by import in folder "enrollment_status_histories"
        Then students were upserted successfully by import

        Examples:
            | condition                                             |
            | enrollment status potential                           |
            | enrollment status non-potential                       |
            | enrollment status enrolled                            |
            | enrollment status graduate                            |
            | enrollment status withdraw                            |
            | enrollment status temporary and potential             |
            | enrollment status enrolled with start date in future  |
            | enrollment status withdrawn with start date in future |
            | enrollment status graduated with start date in future |
            | enrollment status loa with start date in future       |
            | multiple locations and 1 enrollment status            |

    Scenario: Update to "enrollment status temporary" histories by import unsuccessfully
        Given school admin create 1 students with "all fields" by import in folder "students"
        When school admin update 1 students with "changing to enrollment status temporary" by import
        Then student were updated unsuccessfully by import with code "40004" and field "enrollment_status" at row "2"

    Scenario Outline: Create students by import csv file with "<condition>" unsuccessfully
        When school admin create 1 students with "<condition>" by import in folder "enrollment_status_histories"
        Then student were created unsuccessfully by import with code "40004" and field "<field>" at row "2"

        Examples:
            | condition                                                             | field             |
            | enrollment status temporary only                                      | enrollment_status |
            | 1 enrollment status but empty location                                | location          |
            | 2 enrollment status but 1 location                                    | location          |
            | enrollment status temporary with start date in future                 | status_start_date |
            | enrollment status potential with start date in future                 | status_start_date |
            | enrollment status non-potential with start date in future             | status_start_date |
            | enrollment status potential with another status has value out of enum | enrollment_status |
            | enrollment status with value out of enum                              | enrollment_status |

    Scenario Outline: Update students by import csv file with "<condition>" to "<desired-enrollment_status>" unsuccessfully
        Given school admin create 1 students with "<condition>" by import in folder "enrollment_status_histories"
        When school admin update 1 students with "<desired-enrollment_status>" by import
        Then student were updated unsuccessfully by import with code "40004" and field "<field>" at row "2"

        Examples:
            | condition                                 | desired-enrollment_status                                                  | field             |
            | enrollment status potential               | adding new location with enrollment status potential with feature date     | status_start_date |
            | enrollment status potential               | adding new location with enrollment status temporary with feature date     | status_start_date |
            | enrollment status potential               | adding new location with enrollment status non-potential with feature date | status_start_date |

    Scenario Outline: Update students by import csv file with "<condition>" to "<desired-enrollment_status>" successfully
        Given school admin create 1 students with "<condition>" by import in folder "enrollment_status_histories"
        When school admin update 1 students with "<desired-enrollment_status>" by import
        Then students were upserted successfully by import

        Examples:
            | condition                                 | desired-enrollment_status                                              |
            | enrollment status potential               | changing to enrollment status graduate and new date                    |
            | enrollment status potential               | changing to enrollment status withdraw and new date                    |
            | enrollment status potential               | changing to enrollment status LOA and new date                         |
            | enrollment status potential               | changing to enrollment status non-potential and new date               |
            | enrollment status potential               | changing to enrollment status enrolled new date                        |
            | enrollment status enrolled                | changing to enrollment status potential and new date                   |
            | enrollment status enrolled                | changing to enrollment status withdraw and new date                    |
            | enrollment status enrolled                | changing to enrollment status graduate and new date                    |
            | enrollment status enrolled                | changing to enrollment status LOA and new date                         |
            | enrollment status temporary and potential | changing temporary status to enrollment status potential               |
            | enrollment status enrolled                | changing to enrollment status non-potential                            |
            | enrollment status potential               | adding new location with enrollment status potential with current date |
