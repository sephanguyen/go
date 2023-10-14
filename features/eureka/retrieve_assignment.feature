Feature: List assignments

    Scenario: List assignments
        Given some assignments in db
        When user list assignments by ids
        Then returns "OK" status code
        And eureka must return assignments correctly