Feature: Create flashcard content

    Background:
        Given <flashcard>a signed in "school admin"
        And <flashcard>a valid book content
        And user inserts a flashcard

    Scenario Outline: authentication when creating flashcard content
        Given <flashcard>a signed in "<role>"
        When user create a flashcard content
        Then <flashcard>returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | OK               |
            | teacher        | OK               |
            | hq staff       | OK               |
            | student        | PermissionDenied |
            | parent         | PermissionDenied |


    Scenario Outline: create flashcard content with language config
        Given <flashcard>a signed in "school admin"
        When user create a flashcard content with "<language config>"
        Then <flashcard>returns "<status code>" status code

        Examples:
            | language config               | status code |
            | FLASHCARD_LANGUAGE_CONFIG_ENG | OK          |
            | FLASHCARD_LANGUAGE_CONFIG_JP  | OK          |
            | LANGUAGE_CONFIG_ENG           | OK          |
            | LANGUAGE_CONFIG_JP            | OK          |