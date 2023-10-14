Feature: List To Do Item

    Background:
        Given <study_plan> school admin and student login
        And <study_plan> a valid book content
        And student is assigned some valid study plans

    Scenario Outline: authenticate when list to do item
        Given <study_plan> a signed in "<role>"
        When user try list to do item
        Then <study_plan>returns "<msg>" status code

        Examples:
            | role           | msg              | status |
            | school admin   | PermissionDenied | active |
            | admin          | PermissionDenied | active |
            | teacher        | PermissionDenied | active |
            | student        | OK               | active |
            | hq staff       | PermissionDenied | active |
            | center lead    | PermissionDenied | active |
            | center manager | PermissionDenied | active |
            | center staff   | PermissionDenied | active |
            | lead teacher   | PermissionDenied | active |

    Scenario: list student to do item
        Given <study_plan> a signed in "student"
        When user list "<status>" to do item
        Then <study_plan>returns "OK" status code
        And our system must return list to do item correctly

        Examples:
            | status    |
            | active    |
            | completed |
            | overdue   |