Feature: List to do item structured book tree

    Background:
        Given <study_plan> school admin and student login
        And <study_plan> a valid book content
        And valid course and study plan in database

    Scenario Outline: authenticate when list to do item structured book tree
        Given <study_plan> a signed in "<role>"
        When user list to do item structured book tree
        Then <study_plan>returns "<msg>" status code

        Examples:
            | role           | msg              |
            | teacher        | OK               |
            | school admin   | OK               |
            | parent         | PermissionDenied |
            | student        | PermissionDenied |
            | hq staff       | OK               |
            | centre lead    | PermissionDenied |
            | centre manager | PermissionDenied |
            | teacher lead   | PermissionDenied |
    Scenario Outline: authenticate when list to do item structured book tree
        Given <study_plan> a signed in "teacher"
        When user list to do item structured book tree
        Then <study_plan>returns "OK" status code
        And our system must return data correctly