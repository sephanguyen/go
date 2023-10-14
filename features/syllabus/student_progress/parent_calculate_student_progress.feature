Feature: Calculate student progress

    Background:
        Given <student_progress>school admin, parent, teacher and student login
        And "school admin" has created a book with each 3 los, 2 assignments, 4 topics, 3 chapters, 5 quizzes
        And <student_progress>study plan assign to student
        And <student_progress>individual study plan created

    Scenario Outline: Student done some los/assignments correctly
        Given "student" do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
        When parent calculate student progress
        Then <student_progress>returns "OK" status code
        And topic score is "<topic_score>" and chapter score is "<chapter_score>"
        And correct lo completed with "<done_los>" and "<done_assignments>"
        And our system must return learning material result and book tree correctly

        Examples:
            | done_los | correct_quizzes | done_assignments | assignment_mark | topic_score | chapter_score | skipped_topics |
            | 2        | 3               | 1                | 5               | 57          | 57            | 0              |
            | 3        | 5               | 2                | 5               | 80          | 80            | 0              |
            | 0        | 0               | 0                | 0               | 0           | 0             | 0              |