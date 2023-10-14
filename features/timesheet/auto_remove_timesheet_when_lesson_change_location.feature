Feature: Auto create and remove timesheet when lesson update location

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

  Scenario: Timesheet is create and removed when update lesson location
    Given user signed in as school admin
    And user create a lesson in lessonmgmt
    And "1" timesheet will be "created"
    And timesheet have status "Draft"
    And "1" timesheet lesson hours will be "created"
    Then the user updates the lesson location to a different location
    And the lesson "location" was updated
    And timesheet with old location was removed and timesheet with new location was created