@quarantined
Feature: List student by course feature

    Background: valid course student background
        Given a valid course student background
        
    Scenario: List student by course
        Given a signed in "teacher"
        When user list student by course
        Then returns "OK" status code
            And eureka must return correct list of basic profile of students
    Scenario: List student by course when have paging
        Given a signed in "teacher"
        When user list student by course two times with paging
        Then returns "OK" status code
            And eureka must return correct list of basic profile of students
    Scenario: List student by course by name
        Given a signed in "teacher"
            And a Japanese student
        When user list student by course with search_text and paging
        Then eureka must return correct list of basic profile of students
