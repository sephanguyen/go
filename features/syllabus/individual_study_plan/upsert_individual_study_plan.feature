Feature: Upsert individual study plan
    Background:a valid book content
        Given <individual_study_plan>a signed in "school admin"
        And <individual_study_plan>a valid book content

    Scenario Outline: authenticate <role> insert individual study plan
        Given <individual_study_plan>a signed in "<role>"
        When admin insert individual study plan
        Then <individual_study_plan>returns "<msg>" status code
        Examples:
            | role           | msg              |
            | parent         | PermissionDenied |
            | student        | PermissionDenied |
            # | hq staff       | PermissionDenied |
            | centre lead    | PermissionDenied |
            | centre manager | PermissionDenied |
            | teacher lead   | PermissionDenied |
            # | teacher        | PermissionDenied |

    Scenario: admin insert individual study plan
        Given <individual_study_plan>a signed in "<role>"
        And there is a flashcard existed in topic
        And a valid study plan
        Then admin insert individual study plan
        Then <individual_study_plan>returns "<msg>" status code
        And our system stores individual study plan correctly
        Then admin update start date for individual study plan
        And our system updates start date for individual study plan correctly
        Examples:
            | role           | msg |
            | school admin   | OK  |
