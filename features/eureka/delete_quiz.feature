Feature: Delete the quiz
    Background: given a quizet of an learning objective
        Given a quizset with "3" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic

    Scenario: unauthenticated user try to delete the quiz
        Given an invalid authentication token
        When user delete a quiz "1"
        Then returns "Unauthenticated" status code

    Scenario Outline: student, parent and teacher do not have permission to delete the quiz
        Given a signed in "<role>"
        When user delete a quiz "1"
        Then returns "PermissionDenied" status code

        Examples:
            | role           |
            | student        |
            | parent         |
            | teacher        |
            | center lead    |
            | center manager |
            | center staff   |

    Scenario Outline: admin and hq staff try to delete the quiz
        Given a signed in "<role>"
        When user delete a quiz "1"
        Then returns "OK" status code
        And there is no quizset that contains deleted quiz

        Examples:
            | role         |
            | admin        |
            | school admin |
            | hq staff     |

    Scenario: admin missing quiz id when delete quiz
        Given a signed in "school admin"
        When user delete a quiz without quiz id
        Then returns "InvalidArgument" status code
