Feature: List lesson converation
  Background: lesson conversation background
	Given resource path of school "Manabie" is applied
    And a lesson conversation with "1" teachers and "1" students

  Scenario: student list conversation with when only have lesson conversation
    When student "in lesson" call ConversationList
    Then returns "OK" status code
    And tom must not return lesson conversation

