Feature: write comment for student
    In write comment for student
    Teacher upsert comment for student

    Scenario Outline: a student write comment for a student
        Given a signed in "teacher"
        And a valid upsert student comment request with "<row condition>"
        When upsert comment for student
        Then returns "OK" status code
        And BobDB must "<action>" comment for student

    Examples:
        | row condition    | action |
        | new comment      | store  |
        | existing comment | update |

    Scenario Outline: invalid role write comment for a student
        Given a signed in "<signed-in-user>"
        And a valid upsert student comment request with "new comment"
        When upsert comment for student
        Then returns "<status>" status code

    Examples:
        | signed-in-user | status           |
        | teacher        | OK               |
        | school admin   | OK               |
        | parent         | PermissionDenied |
        | student        | PermissionDenied |
    

