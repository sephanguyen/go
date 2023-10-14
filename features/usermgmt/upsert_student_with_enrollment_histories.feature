Feature: Upsert Student With Enrollment Status History

    Background: Sign in with role "staff granted role school admin"
        Given a signed in "staff granted role school admin"

    Scenario Outline: Create a student with enrollment status history and enrollment status is "<condition>"
        When school admin create a student with "enrollment status history" and "<condition>" by GRPC
        Then students were upserted successfully by GRPC
        Examples:
            | condition                       |
            | potential and temporary status  |
            | potential and enrolled status   |
            | potential and withdrawal status |

    Scenario Outline: Create a student with enrollment status history and enrollment status is "<condition>"
        When school admin create a student with "enrollment status history" and "<condition>" by GRPC
        Then students were upserted unsuccessfully by GRPC with "40004" code and "<errorField>" field
        Examples:
            | condition                                    | errorField |
            | potential on future start date               | start_date |
            | non potential on future start date           | start_date |
            | temporary on future date                     | start_date |
            | temporary with start date less than end date | end_date   |