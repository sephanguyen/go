 Feature: display all student that pending reallocation
    
    Background:
        When enter a school
        Given have some centers
        And have some teacher accounts
        And have some student accounts
        And have some courses
        And have some student subscriptions v2
        And have some medias

    Scenario: display all student that pending reallocation
        Given user signed in as school admin
        And some student assigned with reallocate status
        When user retrieve all students that pending reallocation
        Then returns "OK" status code
        And return all student reallocate correctly