Feature: Student Study Plans By Course Id

    Background: some students have study plans for a courses.
        Given <student_study_plan>some valid study plans
        And some students register to the course
        And a user inserted some student study plans to database

    Scenario: StudentStudyPlansManyV2
        When user call StudentStudyPlansManyV2
        Then our system return student study plans correctly

    Scenario: StudentStudyPlansByCourseId
        Given <student_study_plan>there are study plan items existed in study plan
        And <student_study_plan>there are lo study plan items existed in study plan items
        And <student_study_plan>there are assignment study plan items existed in study plan items
        When user call StudentStudyPlansByCourseId
        Then our system return student study plan correctly