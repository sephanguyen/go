Feature: update school date study plan item
  Background: prepare content book and studyplan belongs to 1 student
    Given "school admin" logins "CMS"
    And "student" logins "Learner App"
    And "school admin" has created a content book
    And "school admin" has created a studyplan exact match with the book content for student

  Scenario: Update school date with valid request
    Given update school date with valid request
    Then return successful with updated record with school date "not null"

  Scenario: Update school date with missing study plan item ids
    Given update school date with missing study plan item ids
    Then return error "InvalidArgument"

  Scenario: Update school date with missing student id
    Given update school date with missing student id
    Then return error "InvalidArgument"

  Scenario: Update school date with missing school date
    Given update school date with missing school date
    Then return successful with updated record with school date "null"
