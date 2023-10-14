@runsequence
Feature: Create teacher account

  Background:
    Given "school admin" logins CMS
    And "teacher" logins Teacher App

  Scenario: Create teacher account
    When school admin creates a teacher
    Then school admin sees newly created teacher on CMS
    And teacher logins Teacher App successfully after forgot password