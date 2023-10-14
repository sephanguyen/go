Feature: StudentStudyPlan repository

    Background: some students have study plans for a courses.
        Given <student_study_plan>some valid study plans
        And some students register to the course
        And a user inserted some student study plans to database

    Scenario: FindStudentStudyPlanWithCourseIDs
        When user call FindStudentStudyPlanWithCourseIDs
        Then return a student study plan
