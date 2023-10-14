Feature: Insert a exam LO

    Background:a valid book content
        Given <exam_lo>a signed in "school admin"
        And <exam_lo>a valid book content

    Scenario Outline: Authentication <role> for insert exam LO
        Given <exam_lo>a signed in "<role>"
        When user insert a valid exam LO
        Then <exam_lo>returns "<status>" status code

        Examples:
            | role           | status           |
            | school admin   | OK               |
            | student        | PermissionDenied |
            | parent         | PermissionDenied |
            | teacher        | PermissionDenied |
            | hq staff       | OK               |
            | center lead    | PermissionDenied |
            | center manager | OK               |
            | center staff   | PermissionDenied |

    Scenario: admin create an exam LO in an existed topic
        Given there are exam LOs existed in topic
        When user insert a valid exam LO
        Then our system generates a correct display order for exam LO
        And our system updates topic LODisplayOrderCounter correctly with new exam LO

    Scenario Outline: admin insert exam LO
        When user insert a exam LO without "<field>"
        Then <exam_lo>returns "OK" status code
        And our system must create exam LO with "<field>" as default value

        Examples:
            | field           |
            | instruction     |
            | grade_to_pass   |
            | manual_grading  |
            | time_limit      |
            | maximum_attempt |
            | approve_grading |
            | grade_capping   |
            | review_option   |
