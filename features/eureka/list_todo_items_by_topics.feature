Feature: List todo items by topics
    Background:
        Given a valid "teacher" token
        And a valid course student background

    Scenario: List todo items by topics
        Given user add leaning objectives "<type>" to book
        And user create a valid study plan
        When user get list todo items by topics
        Then returns "OK" status code
        And returns todo items have order correctly
        Examples:
            | type                              |
            | learning_objective                |
            | assignment                        |
            | learning_objective and assignment |

    Scenario: List todo items by topics with null study plan id
        When user get list todo items by topics with null study plan id
        Then returns "InvalidArgument" status code

    Scenario: List todo items by topics with invalid study plan id
        When user get list todo items by topics with invalid study plan id
        Then returns "NotFound" status code

    Scenario: List to do item when delete a learning objective
        Given user add a leaning objective and an assignment with same topic to book
        And user create a valid study plan
        And user delete a "<lo_type>" in book
        When user get list todo items by topics
        Then returns "OK" status code
        And returns todo items have order correctly
        Examples:
            | lo_type            |
            | learning_objective |
            | assignment         |

    Scenario: List todo items by topics with valid available dates
        Given user add leaning objectives "<type>" to book
        And user create a valid study plan
        And update available dates for study plan items
        When user get list todo items by topics with available dates
        Then returns "OK" status code
        And returns todo items have order correctly
        Examples:
            | type                              |
            | learning_objective                |
            | assignment                        |
            | learning_objective and assignment |