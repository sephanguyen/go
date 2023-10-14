Feature: Chat
    Background: Default resource path
        Given resource path of school "Manabie" is applied
    Scenario: unauthenticated user try to send message
        Given a invalid "student" token
        And a invalid conversation_id
        And a SendMessageRequest
        When a "student" send a chat message to conversation
        Then returns "Unauthenticated" status code

    Scenario: user try to send message to a conversation not exist
        Given a valid "student" token
        And a invalid conversation_id
        And a SendMessageRequest
        When a "student" send a chat message to conversation
        Then returns "NotFound" status code

    Scenario Outline: Does not return system message in read API
        Given a list of messages with types "<types>"
        When client calling ConversationDetail
        Then response does not include system message
        Examples:
            | types                                                               |
            | CODES_MESSAGE_TYPE_JOINED_LESSON                                    |
            | CODES_MESSAGE_TYPE_JOINED_LESSON,CODES_MESSAGE_TYPE_END_LIVE_LESSON |


