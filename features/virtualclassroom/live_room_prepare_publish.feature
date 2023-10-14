Feature: Prepare Publish an Uploading Stream in the Live Room

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts

  Scenario: Student prepares to publish and tries again but gets prepared before status
    Given "student" signin system
    And returns "OK" status code
    And user joins a new live room
    And "student" receives channel and room ID and other tokens

    When user prepares to publish in the live room
    Then returns "OK" status code
    And user gets "none" publish status in the live room
    And current live room "includes" streaming learner and gets "1" streaming count

    When user prepares to publish in the live room
    Then returns "OK" status code
    And user gets "prepared before" publish status in the live room
    And current live room "includes" streaming learner and gets "1" streaming count

  Scenario: Student prepares to publish but has reached max limit
    Given "student" signin system
    And returns "OK" status code
    And user joins a new live room
    And "student" receives channel and room ID and other tokens
    And current live room has max streaming learner

    When user prepares to publish in the live room
    Then returns "OK" status code
    And user gets "max limit" publish status in the live room
    And current live room "does not include" streaming learner and gets "13" streaming count

  Scenario: Student prepares to publish with an existing live room state
    Given "teacher" signin system
    And returns "OK" status code
    And user joins a new live room
    And "teacher" receives channel and room ID and other tokens
    And user zoom whiteboard in the live room
    And returns "OK" status code
    
    Given "student" signin system
    And returns "OK" status code
    And user joins an existing live room

    When user prepares to publish in the live room
    Then returns "OK" status code
    And user gets "none" publish status in the live room
    And current live room "includes" streaming learner and gets "1" streaming count

    Given "student" signin system
    And returns "OK" status code
    And user joins an existing live room

    When user prepares to publish in the live room
    Then returns "OK" status code
    And user gets "none" publish status in the live room
    And current live room "includes" streaming learner and gets "2" streaming count

    Given "student" signin system
    And returns "OK" status code
    And user joins an existing live room

    When user prepares to publish in the live room
    Then returns "OK" status code
    And user gets "none" publish status in the live room
    And current live room "includes" streaming learner and gets "3" streaming count