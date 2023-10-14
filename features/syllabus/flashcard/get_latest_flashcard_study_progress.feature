Feature: Retrieve last flashcard study progress

    Background: given a quizet of an learning objective
        Given <flashcard>a signed in "school admin"
        And <flashcard>a valid book content
        And a valid flashcard with quizzes
        And <flashcard>a signed in "student"
        And <flashcard>a course and study plan with "current" student
        And user create some flashcard studies


    Scenario Outline: get latest flashcard study with roles
        Given <flashcard>a signed in "<role>"
        When user get latest flashcard study progress
        Then <flashcard>returns "<msg>" status code
        Examples:
            | role           | msg              |
            | parent         | PermissionDenied |
            | student        | OK               |
            | school admin   | OK               |
            | hq staff       | OK               |
            | teacher        | OK               |

    Scenario: get latest flashcard study with roles
        Given <flashcard>a signed in "student"
        And <flashcard>a course and study plan with "current" student
        And user create some flashcard studies
        When user get latest flashcard study progress
        Then <flashcard>returns "OK" status code
        And returns latest flashcard study progress correctly