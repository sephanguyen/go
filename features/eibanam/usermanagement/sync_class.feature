Feature: Verify course-class after sync data

    Background:
        Given system has synced teacher from partner
        And system has synced course from partner
        And "teacher" logins Teacher App

    # Check class on To-review/Study Plan
    Scenario Outline: View class after sync
        When system syncs class to course which "<status>" current academic year
        Then teacher sees class available in class filter on Teacher App
        And teacher "<result>" the course on Teacher App
        Examples:
            | status | result |
            | in     | sees   |
    #| out of | does not see | Because the ticket https://manabie.atlassian.net/browse/LT-4676

    # # Check class on To-review/Study Plan
    # Scenario: Delete class by sync data
    #     When system syncs class with deleted class
    #     Then teacher does not see the class in class filter on Teacher App

    # Scenario: Edit class name by sync data
    #     When system syncs class with edited class name
    #     Then teacher sees edited class name on Teacher App