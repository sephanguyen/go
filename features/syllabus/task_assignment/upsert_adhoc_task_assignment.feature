Feature: Upsert adhoc task assignment
    Background: a valid course
        Given <task_assignment>a signed in "school admin"
        And <task_assignment>a valid course

    Scenario: Use upsert adhoc task assignment with roles
        Given <task_assignment>a signed in "<role>"
        When user creates a valid adhoc task assignment
        Then <task_assignment>returns "<status code>" status code
        Examples:
            | role         | status code      |
            | school admin | PermissionDenied |
            | student      | OK               |
            | parent       | PermissionDenied |
            | teacher      | PermissionDenied |
            | hq staff     | PermissionDenied |
    # | center lead    | PermissionDenied |
    # | center manager | PermissionDenied |
    # | center staff   | PermissionDenied |

    Scenario: Create adhoc task assignment
        Given <task_assignment>a signed in "student"
        When user creates a valid adhoc task assignment
        Then our system creates adhoc task assignment correctly

    Scenario: Update adhoc task assignment
        Given <task_assignment>a signed in "student"
        And user creates a valid adhoc task assignment
        When user updates the adhoc task assignment
        Then our system updates adhoc task assignment correctly
