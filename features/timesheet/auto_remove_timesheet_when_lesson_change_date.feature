Feature: Auto create and remove timesheet when lesson date be updated

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

  Scenario: Timesheet is create and removed when update lesson date
    Given user signed in as school admin
    And user create a lesson in lessonmgmt
    And "1" timesheet will be "created"
    And timesheet have status "Draft"
    And "1" timesheet lesson hours will be "created"
    Then the user updates the lesson date to a different date
    And the lesson "date" was updated
    And timesheet with old date was removed and timesheet with new date was created