Feature: Modify Live Room Session Time

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts

    Scenario: teacher modify live room session time
      Given "teacher" signin system
      And user joins a new live room
      And returns "OK" status code
      And "teacher" receives channel and room ID and other tokens
      When user modifies the session time in the live room
      Then returns "OK" status code
      And user gets session time in the live room
