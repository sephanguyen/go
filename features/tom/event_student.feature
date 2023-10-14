Feature: Event Student
  Background: Default resource path
    Given resource path of school "Manabie" is applied
  Scenario: Account student is created
    Given a EvtUser with message "CreateStudent"
    When yasuo send event EvtUser
    Then student must be in conversation
    And system must send "created" conversation message

  Scenario: Student list conversation after account student is created
    Given student conversation is created
    When student "" call ConversationList
    Then returns "OK" status code
    And return ConversationList must have "type CONVERSATION_STUDENT,latest_message CODES_MESSAGE_TYPE_CREATED_CONVERSATION"