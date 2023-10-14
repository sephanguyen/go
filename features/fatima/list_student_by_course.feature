@quarantined
Feature: List student by course feature

    Background: valid student packages background
        Given some package data in db
        # And some student packages data in db

    Scenario Outline: List student by course
        Given a signed in "<signed-in user>"
        And a list student by course valid request payload with "<row condition>"
        When user list student by course
        Then returns "OK" status code
        And fatima must return correct list of basic profile of students

        Examples:
            | signed-in user | row condition                     |
            | teacher        | invalid course id                 |
            | teacher        | 100 records                       |
            | teacher        | location ids                      |
            | teacher        | paging                            |
            | teacher        | search text with Japanese student |
            | school admin   | invalid course id                 |
            | school admin   | paging                            |
            | school admin   | 100 records                       |
            | school admin   | location ids                      |
            | school admin   | search text with Japanese student |
