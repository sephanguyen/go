Feature: list flashcard

    Background:a valid book content
        Given <flashcard>a signed in "school admin"
        And <flashcard>a valid book content
        And there are flashcards existed in topic

    Scenario Outline: authenticate list flashcard
        Given <flashcard>a signed in "<role>"
        When user list flashcard
        Then <flashcard>returns "<msg>" status code
        Examples:
            | role           | msg |
            | parent         | OK  |
            | student        | OK  |
            | school admin   | OK  |
            | hq staff       | OK  |
            | teacher        | OK  |
            | centre lead    | OK  |
            | centre manager | OK  |
            | teacher lead   | OK  |

    Scenario: list valid flashcards
        When user list flashcard
        Then <flashcard>returns "OK" status code
        And our system must return flashcards correctly

    Scenario: list valid flashcards has a total question
        Given a valid flashcard with quizzes by learning material ids
        When user list flashcard
        And our system must return flashcards has a total question
