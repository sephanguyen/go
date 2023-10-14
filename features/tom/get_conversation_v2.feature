Feature: Get conversation of a user
  This endpoint for endpoint to get basic information of a conversation using its id
  no matter what type the conversation is

  Background: Default resource path
    Given resource path of school "Manabie" is applied

  Scenario: unauthenticated user try get conversation
    And a GetConversationV2Request
    When a user makes GetConversationV2Request with an invalid token
    Then returns "Unauthenticated" status code

  Scenario: does join show join system message
    Given a chat between a student and "1" teachers
    When a teacher makes GetConversationV2Request with "student conversation id"
    And GetConversationV2Response has latestMessage with content "CODES_MESSAGE_TYPE_CREATED_CONVERSATION"

  @blocker
  Scenario: teacher call GetConversationV2 with a student conversation id
    Given a chat between a student and "1" teachers
    And a teacher sends "text" item with content "hello world"
    When a teacher makes GetConversationV2Request with "student conversation id"
    Then returns "OK" status code
    And tom must return conversation with type "CONVERSATION_STUDENT" in GetConversationV2Response
    And GetConversationV2Response has "1" user with role "student" status "active"
    And GetConversationV2Response has "1" user with role "teacher" status "active"
    And GetConversationV2Response has latestMessage with content "hello world"

  Scenario: teacher call GetConversationV2 with a lesson conversation id
    Given a lesson conversation with "1" teachers and "2" students
    And a teacher sends "1" message with content "hello world" to live lesson chat
    When a teacher makes GetConversationV2Request with "lesson conversation id"
    Then returns "OK" status code
    And tom must return conversation with type "CONVERSATION_LESSON" in GetConversationV2Response
    And GetConversationV2Response has "2" user with role "student" status "active"
    And GetConversationV2Response has "1" user with role "teacher" status "active"
    And GetConversationV2Response has latestMessage with content "hello world"

  Scenario: GetConversationV2 return inactive users
    Given a chat between a student and "2" teachers
    When teacher number 2 leaves student chat
    When a teacher makes GetConversationV2Request with "student conversation id"
    Then returns "OK" status code
    And tom must return conversation with type "CONVERSATION_STUDENT" in GetConversationV2Response
    And GetConversationV2Response has "1" user with role "student" status "active"
    And GetConversationV2Response has "1" user with role "teacher" status "active"
    And GetConversationV2Response has "1" user with role "teacher" status "inactive"
