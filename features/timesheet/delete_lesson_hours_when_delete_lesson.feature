Feature: Delete lesson hours when lesson be deleted

  Background:
    And have some centers
    And have some teacher accounts
    And have 2 teacher accounts will be use for update lesson
    And cloned teacher to timesheet db
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have timesheet configuration is on

  Scenario Outline: Timesheet lesson hours be deleted when school admin can delete a lesson
    Given user signed in as school admin
    And an existing lesson in lessonmgmt
    When user deletes a lesson
    Then returns "<resp_status_code>" status code
    And total "<total>" timesheet lesson hours will be "<action>"
    And timesheet will be deleted

    Examples:
      |  total  | resp_status_code  | action  |
      |  1      | OK                | deleted |