Feature: Auto create timesheet when lesson be created

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And cloned teacher to timesheet db
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And have timesheet configuration is on
    
  Scenario Outline: Timesheet is created when lesson created
    Given user signed in as school admin
    When user create a lesson
    Then "<total>" timesheet will be "<timesheet>"
    And timesheet have status "Draft"
    And "<total timesheet lesson hours>" timesheet lesson hours will be "<timesheet lesson hours>"

    Examples:
      | timesheet | timesheet lesson hours | total | total timesheet lesson hours |
      | created   | created                | 1     | 1                            |