@quarantined
Feature: Delete lesson member and lesson report when student update location
  Background:
    Given an existing student
    And some courses with school id
    And some centers
    And a form's config for "individual lesson report" feature with school id
  Scenario Outline: Delete lesson member and lesson report when student update location
    Given a signed in admin
    And an existing "<lesson-type>" lesson of location <location-init>
    And admin create a lesson report
    When assigning "<student-course-state>" course packages and location <location-init> to existing students
    And returns "OK" status code
    And sync student subscription successfully
    And "enable" Unleash feature with feature name "BACKEND_User_StudentManagement_BackOffice_HandleUpdateStudentCourseLocation"
    And edit "<student-course-state>" course package with location <location-update>
    Then returns "OK" status code
    And lesson member state is "<state>"
    And lesson report state is "<state>"

    Examples:
      | student-course-state   | location-init | lesson-type   | location-update | state       |
      | active                 | 1             | future        | 2               | deleted     |
      | future                 | 1             | future        | 2               | deleted     |
      | active                 | 1             | future        | 1               | not deleted |
      | inactive               | 1             | future        | 2               | deleted     |
      | inactive               | 1             | past          | 2               | not deleted |
      | active                 | 1             | past          | 2               | not deleted |
      | future                 | 1             | past          | 2               | not deleted |
