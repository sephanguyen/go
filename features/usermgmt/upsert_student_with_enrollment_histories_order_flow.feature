Feature: Upsert Student With Enrollment Status History In Order Flow

    Background: Sign in with role "staff granted role school admin"
        Given a signed in as "staff granted role school admin" in "kec-demo" organization

    Scenario Outline: Update students by Upsert with "non-erp status" to "<desired-enrollment_status>" unuccessfully in order flow
        Given school admin create a student with "enrollment status history" and "potential and temporary status" by GRPC
        And school admin simulate the "submit enrollment request event to existed location" to 1 student
        When school admin update a student with "<condition>"
        Then students were upserted unsuccessfully by GRPC with "40004" code and "enrollment_status" field
        Examples:
            | condition                                            |
            | changing to enrollment status potential and new date |
            | changing to enrollment status withdraw and new date  |