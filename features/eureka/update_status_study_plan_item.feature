Feature: update status study plan item
  Background: prepare content book and studyplan belongs to 1 student
    Given "school admin" logins "CMS"
    And "student" logins "Learner App"
    And "school admin" has created a content book
    And "school admin" has created a studyplan exact match with the book content for student

  Scenario Outline: Update status study plan item
    Given update status with "<condition>"
    Then return with error "<error>"

    Examples:
        | condition           | error            |
        | valid_request       | null             |
        | update_request      | updated          |
        | missing_ids         | InvalidArgument  |
        | missing_student_id  | InvalidArgument  |
        | missing_status      | InvalidArgument  |