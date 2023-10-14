Feature: List Student Studyplans
    Background:
        Given <study_plan> a signed in "admin"
        And <study_plan> a valid book content

    Scenario Outline: list student study plan
        Given a study plan in course
        And "2" student in a course
        When <study_plan>user list student study plans
        Then our system must return list student study plan correctly

