Feature: User Message
    To receive message from Bob
    I need to subscribe to nats topic

    Background: User with lesson conversation
        Given resource path of school "Manabie" is applied
        # Given a student conversation with 2 teacher

    Scenario: listen on event user device token
        Given a student conversation with 2 teacher
        Given a valid user device token message
        When bob send event upsert user device token
        Then tom must record device token message

    Scenario: upsert on event user device token
        Given a student conversation with 2 teacher
        Given a valid user device token message
        When bob send event upsert user device token
        Then tom must record device token message
        Then tom must update the user device token
        And tom must update conversation correctly

    Scenario Outline: update locations using usermgmt
        Given a student conversation with 2 teacher
        When usermgmt send event "UpdateStudent" with new token and "<location type>" location in db
        Then tom must update conversation location correctly for event "UpdateStudent"
        Examples:
            | event type    | location type |
            | UpdateStudent | new           |
            | UpdateStudent | no            |

    Scenario: update locations using usermgmt EvtUserInfo and deactivate staffs
        Given a valid "school admin" token
        Given a user group with "Teacher" role and "center" location type
        And a chat between a student and "2" teachers with user groups
        When usermgmt send event "UpdateStudent" with new token and "new" location in db
        Then tom must update conversation location correctly for event "UpdateStudent"
        And teachers are deactivated in conversation members
