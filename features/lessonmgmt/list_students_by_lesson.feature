Feature: List students by lesson

  Background:
    Given enter a school
    When have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias

  Scenario: School admin can create a lesson with all required fields
      Given user signed in as teacher
      And an existing lesson 
      When student list students in that lesson
      Then returns "OK" status code
      And returns a list of students


