Feature: Push notification by publish message to nats jet stream

    Background:
        # Given a new "staff granted role school admin" of Manabie organization with default location logged in Back Office
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students with "1" parents
        And school admin creates "1" course
        And school admin add packages data of those courses for each student
        And student and parent login to learner app
        And update user device token to an "valid" device token

    @blocker
    Scenario Outline: the client push <type> notification to student with config permanent save is <value>
        Given client push "<type>" notification to "<target>" with config permanent save is "<value>"
        And wait to "<type>" notification send
        And wait for FCM is sent to target user
        Then "<target>" must be receive notification
        And notification list of "<target>" must be show right data
        And notification bell display "<new noti>" new notification of "<target>"
        Examples:
            | type      | value | target  | new noti |
            | immediate | false | all     | 0        |
            | immediate | false | student | 0        |
            | immediate | false | parent  | 0        |
            | immediate | true  | all     | 1        |
            | immediate | true  | student | 1        |
            | immediate | true  | parent  | 1        |
            | schedule  | true  | all     | 1        |
            | schedule  | true  | student | 1        |
            | schedule  | true  | parent  | 1        |

    Scenario Outline: the client push <type> notification to student with config permanent save is <value>
        Given client push "immediate" notification using generic id of "<target>" with config permanent save is "false"
        And wait to "immediate" notification send
        Then "<target>" must be receive notification
        And notification list of "<target>" must be show right data
        And notification bell display "<new noti>" new notification of "<target>"
        Examples:
            | target  | new noti |
            | student | 0        |
            | parent  | 0        |
