Feature: Get list student attendance 
  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias

  Scenario: Get list student attendance
    Given user signed in as school admin
    And the system already has "5" lessons in the database
    When users get student attendance
    Then returns "OK" status code
    And the list lesson members have returned correctly

  Scenario Outline: Get list student attendance with filter
    Given user signed in as school admin
    And the system already has "5" lessons in the database with student attendance status "<attendance_status>"
    When users get student attendance with filter "<attendance_status>"
    Then returns "OK" status code
    And the list lesson members have returned correctly
    Examples:
      | attendance_status            |
      | STUDENT_ATTEND_STATUS_ABSENT |
      | STUDENT_ATTEND_STATUS_ATTEND |

