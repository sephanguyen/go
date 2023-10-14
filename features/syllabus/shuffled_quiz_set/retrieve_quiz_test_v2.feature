Feature: Retrieve quiz tests
    I'm a teacher, in a study plan i want to get all students's tests info

    Background: given a quizet of an learning objective
        Given a quizset with "1" quizzes using v2

    Scenario Outline: auth create a quiz test
        Given "1" students do test of a study plan
        And get quiz test of a study plan by "<role>"
        Then <shuffled_quiz_set> returns "<status code>" status code
        Examples:
            | role         | status code |
            | school admin | OK          |
            | admin        | OK          |
            | teacher      | OK          |
            | student      | OK          |
            | hq staff     | OK          |

    Scenario Outline: students take a quiz test successfully
        Given "<student number>" students do test of a study plan
        When teacher get quiz test of a study plan
        Then <shuffled_quiz_set> returns "OK" status code
        And "<student number>" quiz tests infor

        Examples:
            | student number |
            | 3              |
            | 10             |
            | 50             |

    Scenario: students missing study plan item identity
        Given "2" students do test of a study plan
        When teacher get quiz test without study plan item identity
        Then <shuffled_quiz_set> returns "InvalidArgument" status code
