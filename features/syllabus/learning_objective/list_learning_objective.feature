Feature: List learning objective

    Background:
        Given <learning_objective>a signed in "school admin"
        And <learning_objective>a valid book content
        And there are learning objectives existed in topic

    Scenario Outline: authenticate when list learning objective
        Given <learning_objective>a signed in "<role>"
        When user list learning objectives
        Then <learning_objective>returns "<status code>" status code

        Examples:
            | role           | status code |
            | school admin   | OK          |
            | admin          | OK          |
            | teacher        | OK          |
            | student        | OK          |
            | hq staff       | OK          |
            | center lead    | OK          |
            | center manager | OK          |
            | center staff   | OK          |
            | lead teacher   | OK          |

    Scenario: list learning objective
        When user list learning objectives
        Then <learning_objective>returns "OK" status code
        And our system must return learning objectives correctly

    Scenario: list learning objective has a total question
        Given a valid learning objective with quizzes by learning material ids
        When user list learning objectives
        And our system must return learning objective has a total question
