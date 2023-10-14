@runsequence
Feature: Edit teacher info

  Background:
    Given "school admin" logins CMS
    And "teacher" logins Teacher App
    And "student" logins Learner App
    And "parent" logins Learner App
    And school admin has created a teacher
    And school admin has created a student with parent info and visible course

  Scenario: Edit teacher name
    When school admin edits teacher name
    Then school admin sees the edited teacher name on CMS
    And teacher sees the edited teacher name on Teacher App
    And student sees the edited teacher name on Learner App
    And parent sees the edited teacher name on Learner App