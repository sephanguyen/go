Feature: Teacher total unread message
    Background: Default resource path
        Given resource path of school "Manabie" is applied

    Scenario: Get total unread message
        Given teacher joined some conversation in school
        And teacher read all messages
        Given student send 1 message to teacher
        When "teacher" get total unread message
        Then returns "OK" status code
        And Tom must returns 1 total unread message

    Scenario: total unread message by locations
        Given a new school is created with location "default"
        And a signed as a teacher
        And locations "loc 1,loc 2" children of location "default" are created
        And student chat "s1" and parent chat "p1" are created with locations "loc 1"
        And teacher joins all conversations in locations "default"
        And chats "s1,p1" each has new message from student or parent
        And Tom must returns "2" total unread message in locations "loc 1"
        And Tom must returns "2" total unread message in locations "default"
        And Tom must returns "0" total unread message in locations "loc 2"


    Scenario: Total unread message does not count for system message
        Given teacher joined some conversation in school
        When "teacher" get total unread message
        Then returns "OK" status code
        And Tom must returns 0 total unread message


    Scenario: Do not get total unread message for lesson chat
        Given a lesson conversation with "0" teachers and "1" students
        And a teacher joins lesson creating new lesson session
        When "teacher" get total unread message
        Then returns "OK" status code
        And Tom must returns 0 total unread message
