Feature: Unpublish an Uploading Stream

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts

  Scenario: Student unpublish and tries again but gets unpublished before status
    Given "student" signin system
    And returns "OK" status code
    And user joins a new live room
    And "student" receives channel and room ID and other tokens

    When user prepares to publish in the live room
    Then returns "OK" status code
    And user gets "none" publish status in the live room
    And current live room "includes" streaming learner and gets "1" streaming count

    When user unpublish in the live room
    Then returns "OK" status code
    And user gets "none" unpublish status in the live room
    And current live room "does not include" streaming learner and gets "0" streaming count

    When user unpublish in the live room
    Then returns "OK" status code
    And user gets "unpublished before" unpublish status in the live room
    And current live room "does not include" streaming learner and gets "0" streaming count
    
  Scenario: Student tries to unpublish but live room state does not yet exist
    Given "student" signin system
    And returns "OK" status code
    And user joins a new live room
    And "student" receives channel and room ID and other tokens

    When user unpublish in the live room
    Then returns "OK" status code
    And user gets "unpublished before" unpublish status in the live room
    And current live room "does not include" streaming learner and gets "0" streaming count

  Scenario: Student unpublish and there's still a streaming learner left
    Given "student" signin system
    And returns "OK" status code
    And user joins a new live room
    And "student" receives channel and room ID and other tokens

    When user prepares to publish in the live room
    Then returns "OK" status code
    And user gets "none" publish status in the live room
    And current live room "includes" streaming learner and gets "1" streaming count

    When "student" signin system
    Then returns "OK" status code
    And user joins an existing live room

    When user prepares to publish in the live room
    Then returns "OK" status code
    And user gets "none" publish status in the live room
    And current live room "includes" streaming learner and gets "2" streaming count

    When user unpublish in the live room
    Then returns "OK" status code
    And user gets "none" unpublish status in the live room
    And current live room "does not include" streaming learner and gets "1" streaming count

    When user unpublish in the live room
    Then returns "OK" status code
    And user gets "unpublished before" unpublish status in the live room
    And current live room "does not include" streaming learner and gets "1" streaming count