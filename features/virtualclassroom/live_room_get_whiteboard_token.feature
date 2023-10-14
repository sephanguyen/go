Feature: Live Room Get Whiteboard Token

   Scenario: Teacher gets whiteboard token for a new channel
    Given "teacher" signin system
    When user gets whiteboard token for a new channel
    Then returns "OK" status code
    And user receives whiteboard token and other channel details

   Scenario: Teacher gets whiteboard token for an existing channel
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    When user gets whiteboard token for an existing channel
    Then returns "OK" status code
    And user receives whiteboard token and other channel details

   Scenario: Teacher gets whiteboard token for an existing channel but no room ID
    Given "teacher" signin system
    And user joins a new live room
    And returns "OK" status code
    And the existing live room has no whiteboard room ID
    When user gets whiteboard token for an existing channel
    Then returns "OK" status code
    And user receives whiteboard token and other channel details
