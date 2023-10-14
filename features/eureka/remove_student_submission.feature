Feature: User remove student submission

    Scenario: Unauthenticated user not allowed to remove student submission
        Given an invalid authentication token
        When teacher remove student submission
        Then a "Unauthenticated" status code is returned

    Scenario: teacher remove student submission
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        And a signed in "teacher"
        When teacher remove student submission
        Then returns "OK" status code
        And our system must delete student submission correctly

    Scenario: teacher remove student submission
        Given some students are assigned some valid study plans
        And student submit their "existed" content assignment "multiple" times for different assignments
        And a signed in "teacher"
        When "teacher" list the submissions
        And teacher remove student submission after list submissions
        Then the response submissions don't contain submissions were deleted

