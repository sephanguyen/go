Feature: Get conversation of a user
  This endpoint for endpoint to get basic information of a conversation using its id
  no matter what type the conversation is

  Background: Default resource path
    Given resource path of school "Manabie" is applied

  Scenario: unauthenticated user try get conversation
    And a GetConversationRequest
    When a user makes GetConversationRequest with an invalid token
    Then returns "Unauthenticated" status code

  Scenario: does join show join system message
    Given a chat between a student and "1" teachers
    When a teacher makes GetConversationRequest with "student conversation id"
    And GetConversationResponse has latestMessage with content "CODES_MESSAGE_TYPE_CREATED_CONVERSATION"

  @blocker
  Scenario: teacher call GetConversation with a student conversation id
    Given a chat between a student and "1" teachers
    And a teacher sends "text" item with content "hello world"
    When a teacher makes GetConversationRequest with "student conversation id"
    Then returns "OK" status code
    And tom must return conversation with type "CONVERSATION_STUDENT" in GetConversationResponse
    And GetConversationResponse has "1" user with role "student" status "active"
    And GetConversationResponse has "1" user with role "teacher" status "active"
    And GetConversationResponse has latestMessage with content "hello world"

  Scenario: teacher call GetConversation with a lesson conversation id
    Given a lesson conversation with "1" teachers and "2" students
    And a teacher sends "1" message with content "hello world" to live lesson chat
    When a teacher makes GetConversationRequest with "lesson conversation id"
    Then returns "OK" status code
    And tom must return conversation with type "CONVERSATION_LESSON" in GetConversationResponse
    And GetConversationResponse has "2" user with role "student" status "active"
    And GetConversationResponse has "1" user with role "teacher" status "active"
    And GetConversationResponse has latestMessage with content "hello world"

  Scenario: GetConversation return inactive users
    Given a chat between a student and "2" teachers
    When teacher number 2 leaves student chat
    When a teacher makes GetConversationRequest with "student conversation id"
    Then returns "OK" status code
    And tom must return conversation with type "CONVERSATION_STUDENT" in GetConversationResponse
    And GetConversationResponse has "1" user with role "student" status "active"
    And GetConversationResponse has "1" user with role "teacher" status "active"
    And GetConversationResponse has "1" user with role "teacher" status "inactive"
