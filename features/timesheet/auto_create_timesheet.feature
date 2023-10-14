Feature: Auto create timesheet when lesson be updated

  Background:
    When enter a school
    Given have some centers
    And have timesheet configuration is on
    And have some teacher accounts
    And have 2 teacher accounts will be use for update lesson
    And cloned teacher to timesheet db
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias

  Scenario Outline: Timesheet is not existing and lesson scheduling status change
    Given user signed in as school admin
    When user create a lesson in lessonmgmt
    And user updates scheduling status in the lesson is "<scheduling status from>"
    And user updates scheduling status in the lesson is "<scheduling status to>"
    And the lesson scheduling status was updated
    Then "<total>" timesheet will be "<timesheet>"
    And "<total timesheet lesson hours>" timesheet lesson hours will be "<timesheet lesson hours>"

    Examples:
      | scheduling status from         | scheduling status to               | timesheet | timesheet lesson hours | total | total timesheet lesson hours |
      | LESSON_SCHEDULING_STATUS_DRAFT | LESSON_SCHEDULING_STATUS_PUBLISHED | created   | created                | 1     | 1                            |

  Scenario Outline: Timesheet is existing and lesson scheduling status change
    Given "<signed-in user>" signin system
    When user config auto create flag "<auto create config value>" for teachers
    And user create a lesson in lessonmgmt
    And user updates scheduling status in the lesson is "<scheduling status from>"
    And user updates scheduling status in the lesson is "<scheduling status to>"
    And user create a lesson in lessonmgmt
    And user updates scheduling status in the lesson is "<scheduling status from>"
    And user updates scheduling status in the lesson is "<scheduling status to>"
    And the lesson scheduling status was updated
    Then "<total>" timesheet will be "<timesheet>"
    And "<total timesheet lesson hours>" timesheet lesson hours will be "<timesheet lesson hours>"
    And current timesheet lesson hours is "<flagON>"

    Examples:
      | signed-in user                  | scheduling status from         | scheduling status to               | timesheet | timesheet lesson hours | total | total timesheet lesson hours | auto create config value | flagON |
      | staff granted role school admin | LESSON_SCHEDULING_STATUS_DRAFT | LESSON_SCHEDULING_STATUS_PUBLISHED | created   | created                | 1     | 2                            | on                       | on     |
      | staff granted role school admin | LESSON_SCHEDULING_STATUS_DRAFT | LESSON_SCHEDULING_STATUS_PUBLISHED | created   | created                | 1     | 2                            | off                      | off    |

  Scenario Outline: Auto create cannot edit timesheet have status different DRAFT and SUBMITTED
    Given user signed in as school admin
    And user create a lesson in lessonmgmt
    And "1" timesheet will be "created"
    And admin change user timesheet status to "<status>"
    When user remove 1 and add 1 teacher in the lesson
    Then the lesson teachers was updated
    And "<number>" timesheet will be "exists"

    Examples:
      | status    | number |
      | DRAFT     | 1      |
      | SUBMITTED | 1      |
      | APPROVED  | 2      |
      | CONFIRMED | 2      |