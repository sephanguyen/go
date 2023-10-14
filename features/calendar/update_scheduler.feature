Feature: Update scheduler

    Scenario Outline: user update scheduler
        Given signed as "school admin" account
        And user has created a scheduler
        When user update scheduler
        Then returns "OK" status code
        And scheduler has been updated to the database