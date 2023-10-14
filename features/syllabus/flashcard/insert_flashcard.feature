Feature: Insert a flashcard

    Background:a valid book content
        Given <flashcard>a signed in "school admin"
        And <flashcard>a valid book content

    Scenario Outline: authenticate <role> insert flashcard
        Given <flashcard>a signed in "<role>"
        When user inserts a flashcard
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

    Scenario: admin create a flashcard in an existed topic
        Given there are flashcards existed in topic
        When user inserts a flashcard
        And our system generates a correct display order for flashcard
        And our system updates topic LODisplayOrderCounter correctly

    Scenario Outline: admin insert a lmsv2 flashcard
        Given <flashcard>a signed in "school admin"
        When user inserts a lmsv2 flashcard
        Then the lmsv2 flashcard is created with correct data
        And <flashcard>returns "OK" status code
