Feature: Upsert question tag type

    Background:
        Given <question_tag_type>a signed in "school admin"
        And a valid csv content with "id"

    Scenario Outline: authenticate when upsert question tag type
        Given <question_tag_type>a signed in "<role>"
        When user insert some question tag types
        Then <question_tag_type>returns "<status code>" status code

        Examples:
            | role           | status code      |
            | school admin   | OK               |
            | teacher        | PermissionDenied |
            | student        | PermissionDenied |
            | hq staff       | PermissionDenied |
            | center lead    | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |
            | teacher lead   | PermissionDenied |

    Scenario: create question tag type
        When user insert some question tag types
        Then <question_tag_type>returns "OK" status code
        And question tag type must be created

    Scenario: update question tag type
        Given user insert some question tag types
        When user update existed question tag types
        Then <question_tag_type>returns "OK" status code
        And question tag type must be updated

    Scenario: create question tag type without id in csv
        Given a valid csv content with "no id"
        When user insert some question tag types
        Then <question_tag_type>returns "OK" status code