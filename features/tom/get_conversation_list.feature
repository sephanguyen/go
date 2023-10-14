Feature: Get conversation list

  Background: default manabie resource path
    Given resource path of school "Manabie" is applied

  Scenario: Student get conversation list
    Given location configurations conversation value "true" existed on DB
    And a student conversation
    And student go to messages on learner app
    Then student can see 1 conversations

    Given location configurations conversation value "false" existed on DB
    And student go to messages on learner app
    Then student can see 0 conversations
