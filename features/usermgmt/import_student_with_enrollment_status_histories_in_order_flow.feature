@blocker
Feature: Import Student Enrollment Status History In Order Flow

    Background: Sign in with role "staff granted role school admin"
        Given a signed in as "staff granted role school admin" in "kec-demo" organization

    Scenario Outline: Create students by Import with "<condition>" unsuccessfully in order flow
        When school admin create 1 students with "<condition>" by import in folder "enrollment_status_histories"
        Then students were upserted successfully by import

        Examples:
            | condition                                 |
            | enrollment status temporary and potential |
            | enrollment status potential               |
            | enrollment status non-potential           |


    Scenario Outline: Create students by Import with "<condition>" unsuccessfully in order flow
        When school admin create 1 students with "<condition>" by import in folder "enrollment_status_histories"
        Then student were created unsuccessfully by import with code "40004" and field "enrollment_status" at row "2"

        Examples:
            | condition                        |
            | enrollment status enrolled       |
            | enrollment status graduate       |
            | enrollment status withdraw       |
            | enrollment status temporary only |

    Scenario Outline: Update students by Import with "enrollment status potential" to "<desired-enrollment_status>" unuccessfully in order flow
        Given school admin create 1 students with "enrollment status potential" by import in folder "enrollment_status_histories"
        When school admin update 1 students with "<desired-enrollment_status>" by import
        Then student were updated unsuccessfully by import with code "40004" and field "enrollment_status" at row "2"

        Examples:
            | desired-enrollment_status                           |
            | changing to enrollment status graduate and new date |
            | changing to enrollment status withdraw and new date |
            | changing to enrollment status LOA and new date      |
            | changing to enrollment status enrolled new date     |

    Scenario Outline: Update students by Import with "non-erp status" to "<desired-enrollment_status>" unuccessfully in order flow
        Given school admin create 1 students with "enrollment status potential" by import in folder "enrollment_status_histories"
        And school admin simulate the "submit enrollment request event to existed location" to 1 student
        When school admin update 1 students with "<desired-enrollment_status>" by import
        Then student were updated unsuccessfully by import with code "40004" and field "enrollment_status" at row "2"

        Examples:
            | desired-enrollment_status                            |
            | changing to enrollment status potential and new date |
            | changing to enrollment status withdraw and new date  |