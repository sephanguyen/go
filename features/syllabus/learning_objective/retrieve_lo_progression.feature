Feature: Retrieve LO Progression

    Background: a course in some valid locations
        Given <learning_objective>a signed in "school admin"
        And <learning_objective>a valid book content
        And <learning_objective>user creates a course and add students into the course
        And <learning_objective>user adds a master study plan with the created book
        And there is exam LO existed in topic

    Scenario Outline: authenticate retrieve lo progression
        Given <learning_objective>a signed in "<role>"
        When student list lo progression
        Then <learning_objective>returns "<msg>" status code
        Examples:
            | role         | msg              |
            | parent       | PermissionDenied |
            | school admin | PermissionDenied |
            | teacher      | PermissionDenied |
            | hq staff     | PermissionDenied |

    Scenario: retrieve lo progression
        Given <learning_objective>a signed in "school admin"
        And <learning_objective>user create quizzes
        And <learning_objective>user create quiz test v2
        And <learning_objective>a signed in "student"
        And student upsert lo progression with 1 answers
        When student list lo progression
        Then <learning_objective>returns "OK" status code
        And there are 1 answers in the response
