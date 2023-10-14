Feature: Update a learning objective

    Background:a valid book content
        Given <learning_objective>a signed in "school admin"
        And <learning_objective>a valid book content
        And there are learning objectives existed in topic

    Scenario Outline: authenticate <role> updates learning objective
        Given <learning_objective>a signed in "<role>"
        When user updates a learning objective
        Then <learning_objective>returns "<msg>" status code
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
    Scenario: admin update a valid learning objective
        When user updates a learning objective
        And our system updates the learning objective correctly
