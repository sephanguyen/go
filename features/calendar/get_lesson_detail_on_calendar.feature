Feature: Get Lesson Detail on Calendar

    Background:
        Given user signed in as school admin 
        When enter a school
        And have some locations
        And have some teacher accounts
        And have some student accounts
        And have some courses
        And have some student subscriptions
        And have some classrooms
        And have some grades
        And have some medias
        
    Scenario: User can get lesson detail on calendar
        Given signed as "school admin" account
        And an existing "individual" lesson
        And returns "OK" status code
        When user get lesson detail on calendar
        Then returns "OK" status code
        And the lesson detail matches lesson created