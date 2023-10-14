Feature: Create Study Plan

  Background:
    Given "school admin" logins "CMS"
    And "teacher" logins "Teacher App"
    And "student" logins "Learner App"

  Scenario: Create study plan and Add assignment after remove items of book content
    Given "school admin" has created a content book
    And "school admin" has created a studyplan exact match with the book content for student
    When remove all items from book
    And upsert assignment into book
    Then study plan items belong to assignments were successfully created

  Scenario: Create study plan and Add assignment after remove items of book empty
    Given "school admin" has created an empty book
    And "school admin" has created a studyplan exact match with the book empty for student
    When upsert assignment into book
    Then study plan items belong to assignments were successfully created

  @quarantined
  Scenario: Create study plan and Add learning objective after remove items of book content
    Given "school admin" has created a content book
    And "school admin" has created a studyplan exact match with the book content for student
    When remove all items from book
    And upsert learning objective into book
    Then study plan items belong to learning objectives were successfully created
    
  @quarantined
  Scenario: Create study plan and Add learning objective after remove items of book empty
    Given "school admin" has created an empty book
    And "school admin" has created a studyplan exact match with the book empty for student
    When upsert learning objective into book
    Then study plan items belong to learning objectives were successfully created