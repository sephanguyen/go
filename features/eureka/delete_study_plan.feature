Feature: Delete studyplan

  Background:
    Given "school admin" logins "CMS"
    And "teacher" logins "Teacher App"
    And "student" logins "Learner App"
    And "school admin" has created a content book
    And "school admin" has created a studyplan exact match with the book content for student

  Scenario: User delete study plans
    Given user fetchs old study plans
    When user deletes selected study plan
    Then user fetchs new study plans
    And Check selected study plan has been absolutely deleted