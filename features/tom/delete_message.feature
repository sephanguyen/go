Feature: Delete message
    Background: Default resource path
        Given resource path of school "Manabie" is applied

    Scenario: Teacher delete message
        Given a chat between a student and "1" teachers
        And a teacher sends "text" item with content "hello world"
        When "teacher" delete "own" message
        Then returns "OK" status code
        And "student" see deleted message in conversation

    Scenario: Student delete message
        Given a chat between a student and "2" teachers
        And a student sends "image" item with content "hello world"
        When "student" delete "own" message
        Then returns "OK" status code
        And "teacher" see deleted message in conversation

    Scenario: Teacher delete student message in lesson chat
        Given a lesson conversation with "0" teachers and "2" students
        And a teacher joins lesson creating new lesson session
        And students join lesson without refreshing lesson session
        When a student sends "1" message with content "hello world" to live lesson chat
        When "teacher" delete "student" message
        Then returns "OK" status code
        And teacher see deleted message in lesson chat

