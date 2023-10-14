Feature: Create flashcard content

    Background: background flashcard content
        Given <flashcard>a signed in "school admin"
        And <flashcard>a valid book content
        And user inserts a flashcard
        And user create a flashcard content
        And <flashcard>returns "OK" status code

    Scenario: regenerate audio
        When regenerate speeches audio link
        Then options and speeches updated correctly