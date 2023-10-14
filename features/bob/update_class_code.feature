@quarantined
Feature: Update class code

    Background:
        Given "staff granted role school admin" signin system
        And a random number
        And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
        And admin inserts schools

        Given a signed in teacher
        And a CreateClassRequest
        And a "valid" schoolId in CreateClassRequest
        And a valid name in CreateClassRequest
        And this school has config "plan_id" is "School", "plan_expired_at" is "2025-06-30 23:59:59", "plan_duration" is 0
        When user create a class
        Then returns "OK" status code
        And Bob must create class from CreateClassRequest

    Scenario Outline: user try to update a class with wrong case
        Given a signed in "<role>" with school name "S1"
        And a UpdateClassCodeRequest with "<class id>" class id
        When user updates a class
        Then returns "<msg>" status code

        Examples:
            | role            | class id | msg              |
            | unauthenticated | valid    | Unauthenticated  |
            | student         | valid    | PermissionDenied |
            | teacher         | valid    | PermissionDenied |
            | teacher         | invalid  | InvalidArgument  |


    Scenario: owner try to update a class
        Given a UpdateClassCodeRequest with "valid" class id
        When user updates a class
        Then returns "OK" status code
        And Bob must update class code

    Scenario Outline: user try to update a class
        Given a signed in "<role>" with school name "S1"
        And a UpdateClassCodeRequest with "<class id>" class id
        When user updates a class
        Then returns "OK" status code
        And Bob must update class code

        Examples:
            | role         | class id |
            | admin        | valid    |
            | school admin | valid    |
