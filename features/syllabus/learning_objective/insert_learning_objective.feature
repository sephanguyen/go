Feature: Insert a learning objective

    Background:a valid book content
        Given <learning_objective>a signed in "school admin"
        And <learning_objective>a valid book content
        And there are learning objectives existed in topic

    Scenario Outline: authenticate <role> insert learning objective
        Given <learning_objective>a signed in "<role>"
        When user inserts a learning objective
        Then <learning_objective>returns "<msg>" status code
        Examples:
            | role           | msg              |
            | parent         | PermissionDenied |
            | student        | PermissionDenied |
            | hq staff       | OK               |
            | teacher        | PermissionDenied |
            | centre lead    | PermissionDenied |
            | centre manager | PermissionDenied |
            | teacher lead   | PermissionDenied |

    Scenario: admin create a learning objective in an existed topic
        When user inserts a learning objective
        And our system generates a correct display order for learning objective
        And our system updates topic display order counter of learning objective correctly

    Scenario Outline: admin insert learning objective with <field>
        When user inserts a learning objective with "<field>"
        Then <learning_objective>returns "OK" status code
        And our system must create LO with "<field>"

        Examples:
            | field          |
            | manual_grading |
