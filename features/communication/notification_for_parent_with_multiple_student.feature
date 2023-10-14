
Feature: Notification for parent with multiple students
  Scenario: Notification bell display correct read notification count
    Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
    And school admin has created 2 students with the same parent
    And school admin creates "random" courses
    And school admin add packages data of those courses for each student
    And parent login to Learner App
    And admin create notification sending to parent of students
    And parent has 2 items in notification list
    And parent has 2 unread notification
    When parent read notification using created notification id
    Then parent has 0 unread notification
