Feature: manage conversation members with internal API
    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations

    Scenario: add new members to existing conversation
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates a students with first name is "AgoraTest" and last name is "AgoraTest"
        And a new staff with role teacher is created
        And waiting for Agora User has been created
        And student create their conversation
        And teacher is added to conversation

    Scenario: remove members from conversation
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates a students with first name is "AgoraTest" and last name is "AgoraTest"
        And a new staff with role teacher is created
        And waiting for Agora User has been created
        And student create their conversation
        And teacher is added to conversation
        And "teacher" is removed from conversation

    Scenario: remove multiple members from conversation
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates a students with first name is "AgoraTest" and last name is "AgoraTest"
        And a new staff with role teacher is created
        And waiting for Agora User has been created
        And student create their conversation
        And teacher is added to conversation
        And "teacher and student" is removed from conversation