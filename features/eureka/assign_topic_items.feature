Feature: Assign topic item

    Background: valid book content
        Given a signed in "school admin"
        And a list of valid topics
        And admin inserts a list of valid topics

    Scenario: group admin try to assign topic item with invalid request
        Given a list of valid learning objectives
        When user try to assign topic items with invalid request
        Then returns "InvalidArgument" status code

    Scenario: group admin try to assign learning objective to topic
        Given a list of valid learning objectives
        When user try to assign topic items
        Then returns "OK" status code
