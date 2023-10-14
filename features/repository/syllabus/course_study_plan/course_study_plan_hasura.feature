Feature: test hasura CourseStudyPlan

    Background:
        Given some valid study plans
        And a user inserted some course study plans to database

    Scenario: CourseStudyPlansList
        When user call CourseStudyPlansList
        Then our system return course study plans correctly

    Scenario: CourseStudyPlansByCourseId
        Given there are study plan items existed in study plan
        And there are lo study plan items existed in study plan items
        And there are assignment study plan items existed in study plan items
        When user call CourseStudyPlansByCourseId
        Then our system return course study plan correctly
