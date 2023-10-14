Feature: Calculate student progress with all learning mateial type

    Background:
        Given <student_progress>school admin, teacher and student login
        And "school admin" has created a book with each 3 los, 2 flashcard, 2 assignment, 2 task assignment, 2 exam los, 4 topics, 3 chapters, 5 quizzes
        And <student_progress>study plan assign to student
        And <student_progress>individual study plan created

    Scenario Outline: Student done some los/assignments correctly
        Given "student" do test and done "<done_los>" los and "<done_flashcard>" flashcards with "<correct_quizzes>" correctly, "<done_assignments>" assignments and "<done_task_assignment>" task assignments with "<assignment_mark>" point, "<done_exam_lo>" with "<exam_lo_mark>" point and skip "<skipped_topics>" topics
        When student calculate student progress
        Then <student_progress>returns "OK" status code
        And topic score is "<topic_score>" and chapter score is "<chapter_score>"
        And correct lo completed with "<done_los>", "<done_flashcard>", "<done_assignments>", "<done_task_assignment>" and "<done_exam_lo>"
        And our system must return learning material result and book tree correctly

        Examples:
            | done_los | done_flashcard | correct_quizzes | done_assignments | done_task_assignment | assignment_mark | done_exam_lo | exam_lo_mark | topic_score | chapter_score | skipped_topics |
            | 1        | 1              | 3               | 1                | 0                    | 5               | 1            | 7            | 60          | 60            | 0              |
            | 2        | 1              | 5               | 1                | 1                    | 5               | 1            | 7            | 78          | 78            | 0              |
            | 3        | 2              | 4               | 2                | 1                    | 6               | 2            | 7            | 72          | 72            | 0              |
            | 0        | 0              | 0               | 0                | 0                    | 0               | 0            | 0            | 0           | 0             | 0              |