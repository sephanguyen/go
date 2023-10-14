Feature: Upsert school date study plan item

    Background: prepare content book and studyplan belongs to 2 student
        Given <individual_study_plan>a signed in "school admin"
        And <individual_study_plan>a valid book content
        And "school admin" has created a study plan exact match with the book content for 2 student
    
    Scenario Outline: Authen update school date
        Given <individual_study_plan>a signed in "<role>"
        When user update school date
        Then <individual_study_plan>returns "<status code>" status code
        Examples:
            | role         | status code      |
            | parent       | PermissionDenied |
            | student      | PermissionDenied |

    Scenario: Update school date 
        Given <individual_study_plan>a signed in "school admin"
        When user update school date
        Then <individual_study_plan>returns "OK" status code
        And our system stores school date correctly
        And our system triggers data to individual study plan table correctly

    