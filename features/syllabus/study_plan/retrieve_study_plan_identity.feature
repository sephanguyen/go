Feature: Retrieve StudyPlanIdentity

    Background:
        Given <master_study_plan> a signed in "school admin"
        And <master_study_plan>a valid book content
        And "school admin" has created a study plan with the book content for student
        And some study plan items for the study plan

    Scenario Outline: authenticate retrieve study plan identity
        Given <master_study_plan> a signed in "<role>"
        When user retrieves study plan identity
        Then <master_study_plan>returns "<msg>" status code

        Examples:
            | role           | msg |
            | school admin   | OK  |
            | parent         | OK  |
            | student        | OK  |
            | teacher        | OK  |
            | hq staff       | OK  |
            | centre lead    | OK  |
            | centre manager | OK  |
            | teacher lead   | OK  |

    Scenario: student retrieves study plan identity
        Given <master_study_plan> a signed in "student"
        When user retrieves study plan identity
        Then <master_study_plan>returns "OK" status code
        And our system return study plan identity correctly
