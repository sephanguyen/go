Feature: Check quiz correctness
    I submit a answer of the quiz test, then system checks the correctness of the answer

    Background: a course in some valid book content
        Given <quiz>a signed in "school admin"
        And <quiz>user creates a valid book content
        And <quiz>user creates a course and add students into the course
        And <quiz>user adds a master study plan with the created book

    Scenario Outline: authenticate when check quiz correctness
        Given <quiz>a signed in "school admin"
        And <quiz>user create a learning material in "learning objective" type
        And user creates a quiz in "multiple choice" type
        And user updates study plan for the learning material
        When <quiz>a signed in "<role>"
        And user starts and submits a "1" answer in "select"
        Then <quiz>returns "<status_code>" status code

        Examples:
            | role           | status_code      |
            | school admin   | PermissionDenied |
            | admin          | PermissionDenied |
            | lead teacher   | PermissionDenied |
            | teacher        | PermissionDenied |
            | student        | OK               |
            | hq staff       | PermissionDenied |
            | center lead    | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |

    Scenario: Student submits a answer of the quiz test
        Given <quiz>a signed in "school admin"
        And <quiz>user create a learning material in "<lm_type>" type
        And user creates a quiz in "<quiz_type>" type
        And user updates study plan for the learning material
        When <quiz>a signed in "student"
        And user starts and submits a "<content>" answer in "<kind>"
        Then <quiz>returns "OK" status code
        And our system returns "<correctness>" and "<is_correct_all>" correctly

        Examples:
            | lm_type            | quiz_type       | content        | kind   | correctness     | is_correct_all |
            | learning objective | multiple choice | 1              | select |                 |                |
            | learning objective | multiple answer | 1,2            | select |                 |                |
            | learning objective | manual input    | 1              | select | true            | true           |
            | learning objective | fill in blank   | a,B,C          | text   | false,true,true | true           |
            | learning objective | fill in blank   | ａ,Ｂ,Ｃ       | text   | false,true,true | true           |
            | learning objective | order           | keyA,keyB,keyC | order  | true,true,true  | true           |
            | flash card         | pair of word    | Mean A         | text   | true            | true           |

    Scenario: Student retry the quiz test
        Given <quiz>a signed in "school admin"
        And <quiz>user create a learning material in "learning objective" type
        And user creates a quiz in "fill in blank" type
        And user updates study plan for the learning material
        When <quiz>a signed in "student"
        And user starts and submits a "a,b,c" answer in "text"
        And user retry and submits a "A,B,C" answer in "text"
        Then <quiz>returns "OK" status code
