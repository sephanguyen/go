@quarantined
Feature: Handle when `fatima` add new student package v2 or toggle student package's status
    Background: courses is assigned study plan
    Given a valid "admin" token
      And 1 students logins "Learner App"
      And "school admin" logins "CMS"
      And a valid course background
      And "school admin" has created a book with each 2 los, 2 assignments, 2 topics, 1 chapters, 5 quizzes
      And "school admin" has created a studyplan exact match with the book content for student

    Scenario Outline: an admin toggle student package status
    Given a student package v2 with "<status>"
    When the admin toggle student package v2 status
    Then our system have to handle student package v2 correctly
    And courseStudentAccessPaths were created for v2
    Examples: 
      | status  | class_id |
      | active  | exists   |
      | active  | none     |
      | inactive| exists   |
      | inactive| none     |
    
    Scenario: an admin add a student package v2 by package id or courses
    When the admin add a new student package v2 with a package or courses
    Then our system have to handle student package v2 correctly
    Examples:
      | class_id |
      | exists   |
      | none     |

    Scenario: an admin add a student package v2 and update location_id 
    When the admin add a student package v2 and update location_id 
    Then our system have to updated course student access paths correctly for v2
