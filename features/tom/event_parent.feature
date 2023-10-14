Feature: Event Parent

  Background: Default resource path
    Given resource path of school "Manabie" is applied

  Scenario: Account parent is assigned to teacher created
    Given a student conversation with 2 teacher
    And a EvtUser with message "CreateParent"
    When yasuo send event EvtUser
    Then returns "OK" status code
    And tom must create conversation for parent
    And all teacher in student conversation must be in parent conversation
    And system must send "created" conversation message

  Scenario: Account student is assigned to teacher created
    Given a student conversation with 2 teacher
    And a EvtUser with message "ParentAssignedToStudent"
    When yasuo send event EvtUser
    Then returns "OK" status code
    And tom must create conversation for parent
    And all teacher in student conversation must be in parent conversation
    And system must send "created" conversation message

  Scenario: Current user receives "user added" system message
    Given a chat between "1" parents and "1" teachers
    And parents are present
    And another parent account is created with event "ParentAssignedToStudent"
    When this parent is added in these chats
    Then current parents receive message "system" with content "CODES_MESSAGE_TYPE_USER_ADDED_TO_CONVERSATION"
    And teachers receive message "system" with content "CODES_MESSAGE_TYPE_USER_ADDED_TO_CONVERSATION"
