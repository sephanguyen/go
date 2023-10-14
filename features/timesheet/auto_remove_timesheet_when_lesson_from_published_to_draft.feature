Feature: Auto remove timesheet when lesson be updated

  Background:
    When enter a school
    Given have some centers
    And have timesheet configuration is on
    And have some teacher accounts
    And cloned teacher to timesheet db
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias

  Scenario Outline: Timesheet is removed when not contains any information
    Given user signed in as school admin
    And user create a lesson in lessonmgmt
    And "1" timesheet will be "created"
    And timesheet have status "Draft"
    And "1" timesheet lesson hours will be "created"
    Then user updates scheduling status in the lesson is "LESSON_SCHEDULING_STATUS_DRAFT"
    And the lesson scheduling status was updated
    And "<total>" timesheet will be "<timesheet>"
    And "<total timesheet lesson hours>" timesheet lesson hours will be "<timesheet lesson hours>"

    Examples:
      | timesheet   | timesheet lesson hours | total | total timesheet lesson hours |
      | not created | not created            | 0     | 0                            |

  Scenario Outline: Timesheet is removed one timesheet lesson hours
    Given user signed in as school admin
    And user create a lesson in lessonmgmt
    And "1" timesheet will be "created"
    And timesheet have status "Draft"
    And "1" timesheet lesson hours will be "created"
    And user create a lesson in lessonmgmt
    And "1" timesheet will be "created"
    And timesheet have status "Draft"
    And "2" timesheet lesson hours will be "created"
    Then user updates scheduling status in the lesson is "LESSON_SCHEDULING_STATUS_DRAFT"
    And the lesson scheduling status was updated
    And "<total>" timesheet will be "<timesheet>"
    And timesheet have status "Draft"
    And "<total timesheet lesson hours>" timesheet lesson hours will be "<timesheet lesson hours>"

    Examples:
      | timesheet | timesheet lesson hours | total | total timesheet lesson hours |
      | created   | created                | 1     | 1                            |