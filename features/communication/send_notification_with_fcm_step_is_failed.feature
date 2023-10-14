Feature: staff with granted role send notification

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    Scenario Outline: staff with granted role send a notification to user with FCM step is failed
        When current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        And returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        # And update user device token to an "invalid" device token
        And update user device token to an "invalid" device token with "60" fail rate
        When current staff send notification
        Then notificationmgmt services must send notification to user
        And recipient must not receive the notification through FCM mock
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | random        | random       | all             | random       | random        | random             | 1 min            | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT     | false        |
            | student         | random        | random       | all             | random       | random        | random             | 1 min            | NOTIFICATION_STATUS_SCHEDULED | false        |
            | parent          | random        | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT     | false        |
