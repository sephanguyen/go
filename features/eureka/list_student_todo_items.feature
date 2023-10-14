@quarantined
Feature: List student to do items
    Background:
        Given a valid "teacher" token
        And a valid course student
        And add a valid book with some learning objectives to course
        And user create some valid study plan

    Scenario: List student to do items
        Given user update study plans status to "<status>"
        And update dates of study plan items
        When user retrieve list student todo items with status "active"
        Then returns "OK" status code
        And returns todo items total correctly with status "active"

        Examples:
            | status   |
            | archived |
            | active   |

    Scenario: List student to do items
        Given user update study plan items status to "<status>"
        And update dates of study plan items
        When user retrieve list student todo items with status "active"
        Then returns "OK" status code
        And returns todo items total correctly with status "active"

        Examples:
            | status   |
            | archived |
            | active   |
    
    Scenario: List student to do items overdue
        Given user update study plan items status to "active"
        And update dates of study plan items
        When user retrieve list student todo items with status "overdue"
        Then returns "OK" status code
        And returns todo items total correctly with status "overdue"