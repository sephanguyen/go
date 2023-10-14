Feature: Delete lesson hours when recurring lesson be updated
  Background:
    And have some centers
    And have some teacher accounts
    And have 2 teacher accounts will be use for update lesson
    And cloned teacher to timesheet db
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have timesheet configuration is on
    
  # case: change end_date to delete one lesson
  Scenario Outline: Timesheet lesson hours and timesheet be deleted when user update recurring lesson
    Given user signed in as school admin
    And user have created recurring lesson
    And user changed lesson end date to "<end_date>"
    When user update selected lesson by saving weekly recurrence
    Then returns "<resp_status_code>" status code
    And "<total>" timesheet lesson hours should be "<action>"
    And timesheet will be deleted
    Examples:
      | end_date | total | resp_status_code | action |
      | END_DATE | 1     | OK               | deleted|