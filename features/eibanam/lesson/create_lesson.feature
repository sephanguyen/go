Feature: Create a live lesson with all required fields

  Background:
    Given "school admin" logins CMS
    And "teacher" logins Teacher App
    And "student" logins Learner App

  Scenario: School admin can create a live lesson with all required fields
    When school admin creates a new lesson with all required fields
    Then school admin sees the new lesson on CMS
#    And school admin sees message "You have created a lesson successfully!" on CMS
    And teacher sees the new lesson in respective course on Teacher App
    And student sees the new lesson in lesson list on Learner App