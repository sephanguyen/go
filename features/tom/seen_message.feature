Feature: Seen message
    Background: Student has teacher in conversation
        Given resource path of school "Manabie" is applied
        And a student conversation with 2 teacher

    Scenario: Student seen message
        Given teacher send message to conversation
        When student seen conversation
        Then returns "OK" status code
        And tom must mark messages in conversation as read for student
