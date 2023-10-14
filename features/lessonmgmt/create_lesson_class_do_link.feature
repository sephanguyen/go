Feature: User create lesson with class do link

    Background:
        When enter a school
        Given have some centers
        And have some teacher accounts
        And have some student accounts
        And have some courses
        And have some student subscriptions
        And has a ClassDo account

    Scenario: School admin can create a lesson with class do link
        Given user signed in as school admin
        When user creates a new lesson with class do link
        Then returns "OK" status code
        And the lesson was created in lessonmgmt

    Scenario: School admin can get a lesson with class do link
        Given user signed in as school admin
        When user creates a new lesson with class do link
        Then returns "OK" status code
        When user gets the lesson detail with class do link
        Then returns "OK" status code

