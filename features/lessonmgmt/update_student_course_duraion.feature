@quarantined
Feature: User update student course duration 

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some medias
    
  Scenario Outline: School admin update student course duration 
    Given user signed in as school admin
    And user added course to student
    And an existing lesson
    When user updates student course duration
    Then returns "OK" status code
    And inactive student was removed from lesson

