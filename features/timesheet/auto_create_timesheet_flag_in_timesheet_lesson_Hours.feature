@quarantined
Feature: Update auto create timesheet flag affect to timesheet lesson hours

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
    
    
    
  Scenario Outline: School Admin changed auto create flag of other staff
    Given "staff granted role school admin" signin system
    And The teacher have auto create flag is "<before state>"
    And Admin create a future lesson in lessonmgmt for the teacher
    And "1" timesheet will be "created"
    And timesheet have status "Draft"
    And "1" timesheet lesson hours will be "created" with auto create flag "<before state>"
    When School Admin update teacher auto create flag to "<after state>"
    Then returns "OK" status code
    And flag status changed to "<after state>" v2
    And flag in timesheet lesson hours changed to "<after state>"

    Examples:
      | before state | after state | 
      | on           | off         |
      | off          | on          |