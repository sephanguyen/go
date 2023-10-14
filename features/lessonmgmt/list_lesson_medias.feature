Feature: List medias by lesson

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias

  Scenario: Teacher can get list medias
      Given user signed in as teacher
      And an existing lesson 
      When teacher get medias of lesson
      Then returns "OK" status code
      And the list of media match with response medias


