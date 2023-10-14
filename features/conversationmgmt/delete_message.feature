Feature: test delete latest message
    Scenario: test delete latest message
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates a students with first name is "AgoraTest" and last name is "AgoraTest"
        And a new staff with role teacher is created
        And waiting for Agora User has been created
        And current staff create "1" conversations for students
        And current staff add latest message for student's conversation
        When current staff delete this latest message
        Then latest message of student's conversation should be deleted
