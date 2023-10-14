Feature: Import study plan

    Background: a course in some valid book content
        Given <study_plan> a signed in "school admin"
        And <study_plan>user creates a valid book content
        And <study_plan>user creates a course and add students into the course
        And <study_plan>user adds a master study plan with the created book

    Scenario Outline: authenticate when check quiz correctness
        Given <study_plan> a signed in "school admin"
        And user create a learning material in "learning objective" type
        And <study_plan>"school admin" has created a studyplan for all student
        And user bulk upload csv with above study plan 