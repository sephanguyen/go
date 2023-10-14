Feature: Teacher leaves student conversation
    Background: Background name
        Given resource path of school "Manabie" is applied

    Scenario: teacher leaves conversations
        Given random new conversations created
        And a teacher who joined all conversations
        When teacher leaves some conversations
        Then teacher must not be member of conversations recently left

    Scenario: teacher leaves conversation, conversation disappear from list
        Given a chat between a student and "1" teachers
        When teacher leaves student chat
        Then the conversation that teacher left is "not displayed" in conversation list
        When teacher rejoins student chat
        Then the conversation that teacher left is "displayed" in conversation list

    Scenario: teacher leaves conversation, teacher cannot receive message
        Given a chat between a student and "1" teachers
        And teacher leaves student chat
        And student and teachers are present
        When student sends "text" item with content "Hello world"
        Then teacher who left chat does not receive sent message

    Scenario: teacher leaves conversation and rejoins, teacher can receive message
        Given a chat between a student and "1" teachers
        And teacher leaves student chat
        And teacher rejoins student chat
        And student and teachers are present
        When student sends "text" item with content "Hello world"
        Then teachers receive sent message

    Scenario: teacher leaves conversation, teacher cannot send message
        Given a chat between a student and "1" teachers
        And teacher leaves student chat
        And student and teachers are present
        When teacher who left chat cannot send message

    Scenario: members receive leave conversation system message
        Given a chat between a student and "2" teachers
        And student and teachers are present
        And teacher leaves student chat
        And teacher who left conversation receives leave conversation system message
        And other teachers receive leave conversation system message
        And student receive leave conversation system message

    Scenario: teacher leaves conversation and rejoins, teacher can send message
        Given a chat between a student and "1" teachers
        And teacher leaves student chat
        And teacher rejoins student chat
        And student and teachers are present
        When teacher who left chat sends a message
        Then student receives sent message

    Scenario Outline: teacher leaves conversation and rejoins, teacher can send message
        Given a chat between a student and "1" teachers
        Given random new conversations created
        When teacher leaves student chat and "<invalid_chat_type>" chat he does not join
        And the invalid chat does not record teacher membership
        Examples:
            | invalid_chat_type |
            | existing chat     |
            | non existing chat |
