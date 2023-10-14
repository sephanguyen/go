Feature: Upsert master info

    Background: Create study plan and learning material
        Given <master_study_plan> a signed in "school admin"
        And <master_study_plan> list study plan with learning material

    Scenario Outline: authenticate <role> insert master study plan
        Given <master_study_plan> a signed in "<role>"
        When admin insert master study plan
        Then <master_study_plan>returns "<msg>" status code
        Examples:
            | role           | msg              |
            | parent         | PermissionDenied |
            | student        | PermissionDenied |
            | hq staff       | OK               |
            | centre lead    | PermissionDenied |
            | centre manager | PermissionDenied |
            | teacher lead   | PermissionDenied |
            | teacher        | OK               |

    Scenario Outline: admin insert master study plan
        Given <master_study_plan> a signed in "<role>"
        And <master_study_plan> list study plan with learning material
        Then admin insert master study plan
        Then <master_study_plan>returns "<msg>" status code
        And our system stores master study plan correctly
        Then admin update info of master study plan
        And our system updates start date for master study plan correctly
        Examples:
            | role         | msg |
            | school admin | OK  |
