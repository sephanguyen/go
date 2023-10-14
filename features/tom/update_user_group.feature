Feature: Upsert user group
    Background: User with lesson conversation
        Given resource path of school "Manabie" is applied

    Scenario: update locations using usermgmt EvtUserInfo and deactivate staffs
        Given a valid "school admin" token
        And a user group with "Teacher" role and "center" location type
        And a chat between a student with locations and "2" teachers with user groups
        And update user group with "Teacher" role and "center" location type
        Then teachers are deactivated in conversation members
