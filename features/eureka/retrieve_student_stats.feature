Feature: Retrieve Student Statistics
    In order to get the student statistics
    As a student
    I need to retrieve statistics

    Background:
        And a signed in "school admin"

    Scenario: unauthenticated student try to retrieve stats
        Given an invalid authentication token
        And an other student profile in DB
        When user retrieves student stats
        Then returns "Unauthenticated" status code

    Scenario: student try to retrieve another student stats
        Given a signed in "student"
        And an other student profile in DB
        When user retrieves student stats
        Then returns "OK" status code

    Scenario: student retrieves his student stats
        Given a signed in "student"
        And student finishes "2" unassigned learning objectives
        And a list of learning_objective event logs
        And a student inserts a list of event logs
        When user retrieves student stats
        Then returns "OK" status code
        And total_lo_finished must be "2"
        And total_learning_time must be "1800s"
        And achievement crown "ACHIEVEMENT_CROWN_BRONZE" must be 2
