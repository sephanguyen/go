Feature: Update a flashcard

    Background:a valid book content
        Given <flashcard>a signed in "school admin"
        And <flashcard>a valid book content
        And there are flashcards existed in topic


    Scenario Outline: authenticate <role> update flashcard
        Given <flashcard>a signed in "<role>"
        When user updates a flashcard
        Then <flashcard>returns "<msg>" status code
        Examples:
            | role           | msg              |
            | parent         | PermissionDenied |
            | student        | PermissionDenied |
            | school admin   | OK               |
            | hq staff       | OK               |
            | teacher        | PermissionDenied |
            | centre lead    | PermissionDenied |
            | centre manager | PermissionDenied |
            | teacher lead   | PermissionDenied |

    Scenario: admin update a valid flashcard
        When user updates a flashcard
        And our system updates the flashcard correctly
