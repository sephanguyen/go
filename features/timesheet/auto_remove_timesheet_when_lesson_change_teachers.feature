Feature: Auto create or remove timesheet when lesson teachers be updated

  Background:
    When enter a school
    Given have some centers
    And have timesheet configuration is on
    And have 2 teacher accounts 
    And have 2 teacher accounts will be use for update lesson
    And cloned teacher to timesheet db
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias

  Scenario Outline: Timesheet is create or removed when update teachers in lesson
    Given user signed in as school admin
    And user create a lesson in lessonmgmt
    And "2" timesheet will be "created"
    And timesheet have status "Draft"
    And "2" timesheet lesson hours will be "created"
    Then user "<updates action>" teacher in the lesson
    And the lesson teachers was updated
    And "<total>" timesheet will be "<timesheet>"

    Examples:
      | updates action      | timesheet   | total |
      | remove              | not created | 1     |
      | add                 | created     | 3     |
      | 2 remove and 1 add  | created     | 1     |
