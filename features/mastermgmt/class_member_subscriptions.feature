@blocker
Feature: Sync class member
    Background: 
        Given a random number
        And some centers
        And have some courses
        And have a class

    Scenario: Authenticate user try to add course to student package with valid case
        Given "staff granted role school admin" signin system
        When user add a package by course with student package extra for a student
        Then returns "OK" status code
        And server must store correct class members