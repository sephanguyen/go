Feature: Create two lessons with the same information

  Background:
    Given "school admin" logins CMS
    And "student" logins Learner App
    And school admin has created a live lesson on CMS
    And "teacher T1" logins Teacher App

  Scenario: School admin can create a new lesson with the same information as a lesson created before
    When school admin creates a new lesson with exact information as that lesson created before
    Then school admin sees the new lesson on CMS
    And "student" sees the new lesson in lesson list on Learner App
    And "teacher T1" sees the new lesson in respective course on Teacher App