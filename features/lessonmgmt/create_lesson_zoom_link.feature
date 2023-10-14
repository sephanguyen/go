
Feature: User creates lesson

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And have zoom account owner

  Scenario: School admin can create a lesson with zoom link
    Given user signed in as school admin
    When user creates a new lesson with zoom link
    Then returns "OK" status code
    And the lesson was created in lessonmgmt

