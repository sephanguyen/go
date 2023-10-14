Feature: Import Parent and assign to student with username

    Background: Sign in with role "staff granted role school admin"
        Given a signed in "staff granted role school admin"

    Scenario Outline: Import valid csv file with "<row condition>"
        When "school admin" import 1 parent(s) and assign to 1 student(s) with valid payload having "<row condition>"
        Then the valid parent lines are imported successfully
        And returns "OK" status code

        Examples:
            | row condition                        |
            | available username                   |
            | available username with email format |

    Scenario Outline: Import invalid csv file with "<invalid format>"
        When "school admin" import 1 parent(s) and assign to 1 student(s) with valid payload having "<invalid format>"
        Then the invalid parent lines are returned with error code "<error code>"

        Examples:
            | invalid format                                           | error code              |
            | have 1 invalid row with empty username                   | missingMandatory        |
            | have 1 invalid row with username has spaces              | missingMandatory        |
            | have 1 invalid row with username has special characters  | notFollowParentTemplate |
            | have 1 invalid row with existing username                | alreadyRegisteredRow    |
            | have 1 invalid row with existing username and upper case | alreadyRegisteredRow    |
