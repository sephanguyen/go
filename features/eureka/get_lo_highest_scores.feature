Feature: Get highest lo scores
  Background: Valid content books studyplan for that books
    Given "school admin" logins "CMS"
    And "student" logins "Learner App"
    And "teacher" logins "Teacher App"
    And "school admin" has created a book with each 3 los, 2 assignments, 4 topics, 3 chapters, 5 quizzes
    And "school admin" has created a studyplan exact match with the book content for student
  
  Scenario Outline: Student done some los/assignments correctly
    Given "student" do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    When calculate student progress
    Then topic score is "<topic_score>" and chapter score is "<chapter_score>"

  Examples:
    | done_los | correct_quizzes | done_assignments | assignment_mark | topic_score | chapter_score | skipped_topics |
    | 2        | 3               | 1                | 5               | 57          | 57            | 0              |
  
 
