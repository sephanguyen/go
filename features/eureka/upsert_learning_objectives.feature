Feature: Upsert learning objectives

    Background:a valid book content
        And "school admin" logins "CMS"
        And a valid book content

    Scenario Outline: Authentication when upserting learning objectives
        Given "<user>" logins "CMS"
        When user create learning objectives
        And user update learning objectives
        Then returns "<status code>" status code

        Examples:
            | user           | status code      |
            | admin          | OK               |
            | school admin   | OK               |
            | hq staff       | OK               |
            | teacher        | OK               |
            | student        | PermissionDenied |
            | parent         | PermissionDenied |
            | center lead    | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |

    Scenario: create learning objectives
        When user create learning objectives
        Then learning objectives must be created

    Scenario: update learning objectives
        Given user create learning objectives
        When user update learning objectives
        Then learning objectives must be updated

    Scenario Outline: create learning objectives as exam LO
        When user create 1 learning objectives with default values and type "LEARNING_OBJECTIVE_TYPE_EXAM_LO"
        Then returns "OK" status code
        And learning objectives must be created with "<fields>" as default value

        Examples:
            | fields          |
            | instruction     |
            | grade_to_pass   |
            | manual_grading  |
            | time_limit      |
            | maximum_attempt |
            | approve_grading |
            | grade_capping   |
            | review_option   |
            | vendor_type     |

    Scenario Outline: update exam LO with new fields
        Given user create 1 learning objectives with type "LEARNING_OBJECTIVE_TYPE_EXAM_LO"
        When user update "<fields>" of learning objectives
        Then returns "<status code>" status code
        And "<fields>" of learning objectives must be updated

        Examples:
            | fields          | status code |
            | instruction     | OK          |
            | grade_to_pass   | OK          |
            | manual_grading  | OK          |
            | time_limit      | OK          |
            | maximum_attempt | OK          |
            | approve_grading | OK          |
            | grade_capping   | OK          |
            | review_option   | OK          |
