Feature: Update deleted_at field in delete assigment feature
  Background:
    Given "school admin" logins "CMS"
    And "student" logins "Learner App"
    And "teacher" logins "Teacher App"
    And "school admin" has created a content book
    And "school admin" has created a studyplan exact match with the book content for student
  
  Scenario: Delete assignment
    When school admin delete assignment
    Then assignment was successfully deleted in system
    And study plan items belong to assignment were successfully deleted
