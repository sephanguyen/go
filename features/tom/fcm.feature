Feature: FCM related features 
    Scenario: Teacher sends msg while not present does not receive FCM
        Given resource path of school "Manabie" is applied
        And a chat between a student and "1" teachers
        And teachers device tokens is existed in DB
        And a "student" device token is existed in DB
        And teachers are not present
        And student is not present
        When a teacher sends "text" item with content "Hello world"
        And student receives notification
        And teacher does not receives notification
