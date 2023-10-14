@blocker
Feature: Import Parent and assign to student by other users

    Scenario Outline: Only school admin can import parent and assign to student
        Given a signed in "<signed-in user>"
        When "<signed-in user>" import 1 parent(s) and assign to 1 student(s) with valid payload having "valid rows"
        Then returns "<code>" status code

        Examples:
            | signed-in user | code             |
            | school admin   | OK               |
            | teacher        | PermissionDenied |
            | student        | PermissionDenied |
            | parent         | PermissionDenied |
