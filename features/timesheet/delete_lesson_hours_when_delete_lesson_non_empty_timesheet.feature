Feature: Non-empty timesheet and lesson hours cannot be deleted when lesson be deleted

  Background:
    And have some centers
    And have some teacher accounts
    And have timesheet configuration is on
    And have 2 teacher accounts will be use for update lesson
    And cloned teacher to timesheet db
    And have some student accounts
    And have some courses
    And have some student subscriptions

  Scenario Outline: Non-empty timesheet cannot be deleted when lesson be deleted
    Given user signed in as school admin
    And an existing lesson in lessonmgmt
    And "<other_working_hours_count>" other working hours records be added to existing timesheet
    And "<transportation_expense_count>" transportation expenses be added to existing timesheet
    When user deletes a lesson
    Then returns "<resp_status_code>" status code
    And total "<total>" timesheet lesson hours will be "<timesheet_lesson_hours>"
    And "<other_working_hours_count>" other working hours records are still existing
    And "<transportation_expense_count>" transportation expenses are still existing
    And "<total>" timesheet cannot be deleted

    Examples:
      | other_working_hours_count  | transportation_expense_count | total  | resp_status_code  | timesheet_lesson_hours |
      | 3                          | 1                            | 1      | OK                | deleted                |

  Scenario Outline: Timesheet lesson hours be deleted when user delete lesson recurring
    Given user signed in as school admin
    When user creates recurring lesson
    Then returns "<resp_status_code>" status code
    When user deletes recurring lesson from "<delete_from>" with "<method>"
    Then returns "<resp_status_code>" status code
    And "<deleted_timesheet_lesson_hours>" timesheet lesson hours should be "<action>"
    And "<remaining_timesheet_lesson_hours>" timesheet lesson hours are still existed
    And "<deleted_timesheet_lesson_hours>" timesheet recurring lesson will be deleted

    Examples:
      | method    | delete_from | deleted_timesheet_lesson_hours | remaining_timesheet_lesson_hours | resp_status_code | action  |
      | one_time  | 3           | 1                              | 3                                | OK               | deleted |
      | recurring | 0           | 4                              | 0                                | OK               | deleted |
      | recurring | 2           | 2                              | 2                                | OK               | deleted |
      | recurring | 1           | 3                              | 1                                | OK               | deleted |

  Scenario Outline: Non-empty timesheet cannot be deleted when user delete lesson recurring
    Given user signed in as school admin
    When user creates recurring lesson
    And "<total>" other working hours records be added to timesheet lesson recurring
    And "<total>" transportation expenses be added to timesheet lesson recurring
    When user deletes recurring lesson from "<delete_from>" with "<method>"
    Then returns "<resp_status_code>" status code
    And "<total_deleted>" timesheet lesson hours should be "<action>"
    And "<total_remaining>" timesheet lesson hours are still existed
    And "<total_timesheet>" timesheet cannot be deleted

    Examples:
      | method    | delete_from | total_deleted | total_remaining | resp_status_code | action   | total  | total_timesheet |
      | one_time  | 3           | 1             | 3               | OK               | deleted  | 3      | 4               |
      | recurring | 0           | 4             | 0               | OK               | deleted  | 3      | 4               |