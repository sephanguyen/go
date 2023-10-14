Feature: Store device token
  In order to use app
  As a user
  I need to store my device token

  Scenario: admin try to store device token
    Given a signed in admin
    And a valid device token
    When user try to store device token
    Then Bob must store the user's device token
    And Bob must publish event to user_device_token channel
    And Tom must record new user_device_tokens with device_token info

  Scenario: student try to store device token
    Given a signed in student
    And a valid device token
    When user try to store device token
    Then Bob must store the user's device token
    And Bob must publish event to user_device_token channel
    And Tom must record new user_device_tokens with device_token info