@quarantined
Feature: Retrieve class
    Background:
        Given "staff granted role school admin" signin system
        And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
        And admin inserts schools
        And some class members

    Scenario: teacher retrieve class by ids
        Given "staff granted role teacher" signin system
        When user retrieve class by ids
        Then returns "OK" status code
        And bob must return correct class ids

    Scenario: student retrieve class by ids
        Given "student" signin system
        When user retrieve class by ids
        Then returns "PermissionDenied" status code