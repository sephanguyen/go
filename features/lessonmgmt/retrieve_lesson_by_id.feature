Feature: User get detail lesson

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias

  Scenario Outline: School admin get detail a lesson
    Given user signed in as school admin
    And an existing lesson
    When user get detail lesson 
    Then returns "OK" status code
    And the lesson detail match lesson created