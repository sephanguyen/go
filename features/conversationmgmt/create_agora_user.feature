Feature: test create Agora user with Manabie created user event

    Scenario: happy case create Agora user (student)
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates a students with first name is "AgoraTest" and last name is "AgoraTest"

    Scenario: create Agora user (teacher)
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And a new staff with role teacher is created
        And agora teacher is created successfully