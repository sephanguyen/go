Feature: Update publish status learning materials

    Background:
        Given a signed in "school admin"
        And user adds a simple book content
        And user adds some learning materials to topic of the book

    Scenario: Update publish status learning material to <is_published>
        When user updates publish status of learning material to "<is_published>"
        Then our system must update the publish status of learning material correctly

        Examples:
            | is_published |
            | true         |
            | false        |
