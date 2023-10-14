Feature: Re calculate student progress after delete LO

    Background:
        Given "school admin" logins "CMS"
        And "teacher" logins "Teacher App"
        And "student" logins "Learner App"
        And "school admin" has created a book with each 4 los, 1 assignments, 1 topics, 1 chapters, 5 quizzes
        And "school admin" has created a studyplan exact match with the book content for student
        And "student" do test and done "4" los with "4" correctly and "1" assignments with "10" point and skip "0" topics
        When teacher retrieve student progress
        Then our system have to return student progress correctly
    
    Scenario: Delete lo and re-calculate student progress
        When school admin delete a los
        And teacher retrieve student progress
        Then our system have to return student progress correctly
        And topic learning objectives of deleted los were successfully deleted
