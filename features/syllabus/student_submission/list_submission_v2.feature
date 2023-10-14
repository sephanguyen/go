Feature: List submissions v2
  #retrieve submissions of student mapped with those locations
  Background: students submit assignments
    Given <student_submission> a signed in "school admin"
    And <student_submission> a valid book content
    And <student_submission> some students added to course in some valid locations
    And create a study plan for that course
    And students submit their assignments

  Scenario Outline: authenticate  list submissions v2
    Given <student_submission> a signed in "<role>"
    When user using list submissions ver2 with "true" student name
    Then <student_submission> returns "<msg>" status code

    Examples:
      | role         | msg              |
      | school admin | OK               |
      | parent       | PermissionDenied |
      | student      | PermissionDenied |
      | teacher      | OK               |
      | hq staff     | OK               |

  Scenario Outline: student list submissions v4
    Given <student_submission> a signed in "teacher"
    When user using list submissions ver2 with "<value>" student name
    Then returns list student submissions v2 correctly with "<value>"

    Examples:
      | value |
      | true  |
      | false |
