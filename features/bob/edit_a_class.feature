@quarantined
Feature: Edit a class
    Background:
        Given "staff granted role school admin" signin system
        And a random number
        And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
        And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
        And admin inserts schools

        Given create a class with school name "S1" and expired at "2150-12-12"

    Scenario: unauthenticated user try to edit a class
        Given an invalid authentication token
        And a EditClassRequest with class name is "edit-class-name"
        When user edit a class
        Then returns "Unauthenticated" status code

    Scenario: another teacher try to edit the class not owner
        Given a signed in teacher
        And a EditClassRequest with class name is "edit-class-name"
        And a "valid" classId in EditClassRequest
        When user edit a class
        Then returns "PermissionDenied" status code

    Scenario: owner class try to edit a class
        And a EditClassRequest with class name is "edit-class-name"
        And a "valid" classId in EditClassRequest
        When user edit a class
        Then returns "OK" status code
        And Bob must update class in db
        And Bob must push msg "EditClass" subject "Class.Upserted" to nats

    Scenario Outline: admin try to edit class with wrong case
        Given a signed in "<role>" with school name "<school name>"
        And a EditClassRequest with class name is "class name"
        And a "<class id>" classId in EditClassRequest
        When user edit a class
        Then returns "<err>" status code

        Examples:
            | role         | school name | class name | class id | err              |
            | admin        | S1          | edit name  |          | InvalidArgument  |
            | school admin | S2          | edit name  | valid    | PermissionDenied |


    Scenario Outline: admin try to edit class with valid data
        Given a signed in "<role>" with school name "<school name>"
        And a EditClassRequest with class name is "class name"
        And a "<class id>" classId in EditClassRequest
        When user edit a class
        Then returns "OK" status code
        And Bob must update class in db
        And Bob must push msg "EditClass" subject "Class.Upserted" to nats
        Examples:
            | role         | school name | class name | class id |
            | admin        | S1          | edit name  | valid    |
            | school admin | S1          | edit name  | valid    |
