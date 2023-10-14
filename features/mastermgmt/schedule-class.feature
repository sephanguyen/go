Feature: Schedule Class

    Background: 
        Given a random number
        And some centers
        And have some courses
        And a list of class are existed in DB
        And "staff granted role school admin" signin system
        And user add a package by course with student package extra for a student

    Scenario: admin try to save record to reserve class
        Given user has scheduled class to reserve class
        When user schedule class to reserve class again with other class
        Then returns "OK" status code
        And reserve class must be stored correct on db

        # Fake behavior when cron job trigger apply reserve class to register student class
        When user call func wrapper register student class
        Then returns "OK" status code
        And student package class with reserve class must be stored in database

    Scenario: admin try to retrieve scheduled class info
        Given user has scheduled class to reserve class
        When user retrieve scheduled class info
        Then returns "OK" status code
        # TODO: add test assert data response later
