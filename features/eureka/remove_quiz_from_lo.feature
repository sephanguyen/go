
Feature: remove a quiz from learning objective
    Background: given a quizet of an learning objective
        Given a quizset with "3" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic

    Scenario Outline: authenticate to remove the quiz from lo
        Given a signed in "<role>"
        When user remove a quiz "1" from lo
        Then returns "<status code>" status code
        Examples:
            | role           | status code      |
            | student        | PermissionDenied |
            | teacher        | PermissionDenied |
            | parent         | PermissionDenied |
            | admin          | OK               |
            | school admin   | OK               |
            | hq staff       | OK               |
            | center lead    | PermissionDenied |
            | center manager | PermissionDenied |
            | center staff   | PermissionDenied |

    Scenario: remove the quiz from lo successfully
        Given a signed in "school admin"
        When user remove a quiz "1" from lo
        Then LO does not contain deleted quiz

    Scenario: remove the quiz from lo successfully that not belong to question group
        Given existing question group
        And user upsert a valid "questionGroup" single quiz
        And user upsert a valid "valid" single quiz
        And a signed in "school admin"
        When user remove a quiz "1" from lo
        Then LO does not contain deleted quiz

    Scenario: remove the quiz from lo successfully that belong to question group
        Given existing question group
        And user upsert a valid "questionGroup" single quiz
        And a signed in "school admin"
        When user remove a quiz "1" from lo
        Then LO does not contain deleted quiz

    Scenario Outline: admin missing field when remove quiz
        Given a signed in "school admin"
        When user remove a quiz without "<field>"
        Then returns "<status code>" status code
        Examples:
            | field   | status code     |
            | quiz id | InvalidArgument |
            | lo id   | InvalidArgument |
            | none    | OK              |
