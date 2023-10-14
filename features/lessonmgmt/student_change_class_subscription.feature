Feature: Student Change Class Subscription

    Background:
        Given "staff granted role school admin" signin system
        And have some teacher accounts
        And a list of locations with are existed in DB
        And have some student accounts
        And have some courses
        And have some classes assign to courses
        And have some lessons assign to classes
    Scenario: Student join class
        Given a student add courses
        Then returns "OK" status code
        And student must join to lessons of class
    Scenario: Student leave class
        Given a student add courses
        Then returns "OK" status code
        And student must join to lessons of class
        When student change other class
        Then returns "OK" status code
        And student must join to lessons of class
    Scenario: Student change duration join class
        Given a student add courses
        Then returns "OK" status code
        And student must join to lessons of class
        When student change duration
        Then returns "OK" status code
        And student must leave to lessons have a start time less than class end time duration 
