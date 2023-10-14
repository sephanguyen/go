Feature: update learning material name
    Background:
        Given <learning_material>a signed in "school admin"
        And <learning_material>a valid book content
        And some existing learning materials in an arbitrary topic of the book

    Scenario Outline: authenticate <role> update learning material name
        Given <learning_material>a signed in "<role>"
        When user update LM name
        Then <learning_material>returns "<msg>" status code
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
    # this scenario, use random to make fair, insrease the coverage
    Scenario: update learning material name
        When user update LM name
        Then our system must update learning material name correctly

