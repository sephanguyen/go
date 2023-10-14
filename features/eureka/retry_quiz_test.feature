@quarantined
Feature: Calculate student progress
  Background: Valid content books studyplan for that books
    Given "school admin" logins "CMS"
    And "student" logins "Learner App"
    And "teacher" logins "Teacher App"
    And "school admin" has created a book with each 1 los, 0 assignments, 1 topics, 1 chapters, 5 quizzes
    And "school admin" has created a studyplan exact match with the book content for student

  Scenario Outline: Student done some los/assignments correctly
    Given "student" do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    When calculate student progress
    Then topic score is "<topic_score>" and chapter score is "<chapter_score>"
    And correct lo completed with "<done_los>" and "<done_assignments>"
    Then student retry do wrong quizzes with "<num_quizzes>" correct 
    When calculate student progress
    Then topic score is "<new_topic_score>" and chapter score is "<new_chapter_score>"

  Examples:
      | done_los | correct_quizzes | done_assignments | assignment_mark | topic_score | chapter_score | skipped_topics | num_quizzes | new_topic_score | new_chapter_score |
      | 1        | 1               | 0                | 0               | 20          | 20            | 0              | 4           | 100             | 100               |
      | 1        | 2               | 0                | 0               | 40          | 40            | 0              | 2           | 80              | 80                |
  
 