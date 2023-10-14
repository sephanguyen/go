Feature: Modify Live Room Spotlight

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts

    Scenario: teacher modify live room spotlight then unspotlight
      Given "teacher" signin system
      And user joins a new live room
      And returns "OK" status code
      And "teacher" receives channel and room ID and other tokens
      When user "adds" a spotlighted user in the live room
      Then returns "OK" status code
      And user gets correct spotlight state in the live room

      When user "removes" a spotlighted user in the live room
      Then returns "OK" status code
      And user gets correct spotlight state in the live room
