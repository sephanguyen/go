Feature: Retrieve allocate marker

    Background:
        Given <study_plan> a signed in "teacher"
        And an existing allocate marker

    Scenario Outline: authenticate when <role> retrieve allocate marker
        Given <study_plan> a signed in "<role>"
        When <study_plan>user send a retrieve allocate marker request
        Then <study_plan>returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | OK               |
            | student        | PermissionDenied |
            | hq staff       | OK               |
            | center manager | OK               |
            | lead teacher   | OK               |
            # | center lead    | OK               |
            # | center staff   | OK               |



    Scenario: retrieve allocate marker
        When <study_plan>user send a retrieve allocate marker request
        Then <study_plan>returns "OK" status code
        And our system must return allocate marker correctly
