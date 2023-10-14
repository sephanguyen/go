Feature: Mark system notification as new/done

    Scenario Outline: user set the notification status
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        Given some staffs with random roles and granted organization location of current organization
        And staff create system notification with "5" new and "0" done and "0" unenabled
        And waiting for kafka sync data
        When user set "<status>" the system notification
        Then mark the system notification as status "<status>"
        Examples:
            | status                          |
            | SYSTEM_NOTIFICATION_STATUS_NEW  |
            | SYSTEM_NOTIFICATION_STATUS_DONE |
