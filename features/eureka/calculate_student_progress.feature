Feature: Calculate student progress
  Background: Valid content books studyplan for that books
    Given "school admin" logins "CMS"
    And "student" logins "Learner App"
    And "school admin" has created a book with each 3 los, 2 assignments, 4 topics, 3 chapters, 5 quizzes
    And "school admin" has created a studyplan exact match with the book content for student
    And "teacher" logins "Teacher App"
  
  Scenario Outline: Invalid request
    Given "student" do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    When calculate student progress with missing "<request_param>"
    Then return with error "<error>"

  Examples:
    | done_los | correct_quizzes | done_assignments | assignment_mark | skipped_topics | request_param | error           |
    | 2        | 3               | 1                | 5               | 0              | course_id     | InvalidArgument |
    | 2        | 3               | 1                | 5               | 0              | book_id       | InvalidArgument |
    | 2        | 3               | 1                | 5               | 0              | student_id    | InvalidArgument |

  Scenario Outline: Student done some los/assignments correctly
    Given "student" do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    When calculate student progress
    Then topic score is "<topic_score>" and chapter score is "<chapter_score>"
    And correct lo completed with "<done_los>" and "<done_assignments>"

  Examples:
      | done_los | correct_quizzes | done_assignments | assignment_mark | topic_score | chapter_score | skipped_topics |
      | 2        | 3               | 1                | 5               | 57          | 57            | 0              |
      | 3        | 5               | 2                | 5               | 80          | 80            | 0              |
      | 0        | 0               | 0                | 0               | 0           | 0             | 0              |
  
  Scenario Outline: Student done some los/assignments correctly and skip some topic
    Given "student" do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    When calculate student progress
    Then topic score is "<topic_score>" and chapter score is "<chapter_score>"

  Examples:
      | done_los | correct_quizzes | done_assignments | assignment_mark | topic_score | chapter_score | skipped_topics |
      | 2        | 3               | 1                | 5               | 57          | 57            | 2              |
      | 3        | 5               | 2                | 5               | 80          | 80            | 2              |
      | 0        | 0               | 0                | 0               | 0           | 0             | 2              |
 
  Scenario Outline: Student done some los/assignments correctly and admin deletes topics
    Given "student" do test and done "<done_los_1>" los with "<correct_quizzes_1>" correctly and "<done_assignments_1>" assignments with "<assignment_mark_1>" point in the first two topics and done "<done_los_2>" los with "<correct_quizzes_2>" correctly and "<done_assignments_2>" assignments with "<assignment_mark_2>" point in the other
    When school admin delete "<index_topic_list>" topics
    When calculate student progress
    Then first pair topic score is "<topic_score_1>" and second pair topic score is "<topic_score_2>" and chapter score is "<chapter_score>"

  Examples:
      | done_los_1 | correct_quizzes_1 | done_assignments_1 | assignment_mark_1  | done_los_2 | correct_quizzes_2 | done_assignments_2 | assignment_mark_2 | topic_score_1 | topic_score_2 | chapter_score | index_topic_list |
      | 2          | 3                 | 1                  | 5                  | 3          | 5                 | 2                  | 5                 | 57            | 80             | 72           | 0,4,8            | 
      | 3          | 5                 | 2                  | 5                  | 2          | 3                 | 1                  | 5                 | 80            | 57             | 65           | 1,5,9            |
      | 0          | 0                 | 0                  | 0                  | 0          | 0                 | 0                  | 0                 | 0             | 0              | 0            | 2,6              |

  Scenario Outline: Student done some los/assignments with some of them are task assignment 
    Given some of created assignments are task assignment
    And "student" do test and done "<done_los>" los with "<correct_quizzes>" correctly and "<done_assignments>" assignments with "<assignment_mark>" point and skip "<skipped_topics>" topics
    When calculate student progress
    Then topic score is "<topic_score>" and chapter score is "<chapter_score>"

  Examples:
      | done_los | correct_quizzes | done_assignments | assignment_mark | topic_score | chapter_score | skipped_topics |
      | 0        | 0               | 2                | 6               | 60          | 60            | 0              |
      | 0        | 0               | 0                | 0               | 0           | 0             | 0              |
