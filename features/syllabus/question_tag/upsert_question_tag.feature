Feature: Upsert question tag

    Background:
        Given <question_tag>a signed in "school admin"
        And some question tag types existed in database
        And a valid csv content with some valid question tags

    Scenario Outline: authenticate when upsert question tag
        Given <question_tag>a signed in "<role>"
        When user upsert question tag
        Then <question_tag>returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | OK               |
            | admin          | OK               |
            | teacher        | PermissionDenied |
            | student        | PermissionDenied |
            | hq staff       | PermissionDenied |
            | center lead    | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |
            | lead teacher   | PermissionDenied |

    Scenario: create question tag
        When user create question tag
        Then <question_tag>returns "OK" status code
        And question tag must be created

    Scenario: update question tag
        Given user create question tag
        When user update question tag
        Then <question_tag>returns "OK" status code
        And question tag must be updated
