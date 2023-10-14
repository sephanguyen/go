Feature: School admin creates a new course
  Teacher sees it on Teacher Web
  Learner can not see it on Learner Web

  Background:
    Given "school admin" logins CMS
    And "teacher" logins Teacher App
    And "student" logins Learner App

  Scenario: School admin creates a new course
    Given school admin is on the course page
    When school admin creates a new course
    Then school admin sees the new course on CMS
    And teacher sees the new course on Teacher App
    And student can not see the new course on Learner App