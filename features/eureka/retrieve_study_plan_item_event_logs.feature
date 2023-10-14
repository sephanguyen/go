Feature: Retrieve Study Plan Item Event Logs

    Background:
        And a signed in "teacher"
        And an assigned student
        And some learning_objective student event logs are existed in DB

    Scenario: student try to retrieve study plan item event logs
        Given a signed in "student"
        When retrieve study plan item event logs
        Then returns "PermissionDenied" status code

    Scenario: teacher try to retrieve study plan item event logs
        Given a signed in "teacher"
        When retrieve study plan item event logs
        Then returns "OK" status code
