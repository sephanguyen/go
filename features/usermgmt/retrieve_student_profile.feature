@runsequence
Feature: Retrieve Student Profile

    Scenario: student retrieves student profile with empty query param
        Given a signed in student
        When user retrieves student profile
        Then returns "OK" status code
        And return the requester's profile
    
    Scenario: student retrieves another student profile
        Given a signed in student
        And an other student profile in DB
        When user retrieves student profile
        Then returns "OK" status code
        And returns requested student profile

    Scenario: unauthenticated user retrieves student profile
        Given an invalid authentication token
        And an other student profile in DB
        When user retrieves student profile
        Then returns "Unauthenticated" status code

    Scenario: teacher retrieves student profile
        Given a signed in teacher
        When teacher retrieves a "<kind of student>" student profile
        Then returns "OK" status code
        And returns requested student profile

    Examples:
      | kind of student      |
      | newly created        |
      | has signed in before |
    
    Scenario: teacher retrieves student profile with empty query param
        Given a signed in teacher
        When user retrieves student profile
        Then returns "OK" status code
        And returns empty student profile

    Scenario: teacher retrieves student profile with grade info
        Given a signed in teacher
        And generate grade master
        And student info with grade master request
        When "staff granted role school admin" create new student account
        And teacher retrieves the student profile
        Then returns "OK" status code
        And returns student profile with correct grade info
