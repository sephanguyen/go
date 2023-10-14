Feature: Export academic calendar

    Background:
        Given some centers
        And have academic year

    Scenario: admin try to export academic calendar csv file success
        Given "staff granted role school admin" signin system
        And a valid academic calendar payload
        When user import academic calendar by csv file
        And user try to export academic calendar
        Then returns academic calendar in csv with Ok status code
