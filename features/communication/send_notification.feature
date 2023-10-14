Feature: staff with granted role create and send a notification

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    Scenario: "<staff with granted role>" send notification by role successfully
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        And current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        And update user device token to an "valid" device token
        When current staff send notification
        Then notificationmgmt services must send notification to user
        And wait for FCM is sent to target user
        And recipient must receive the notification through FCM mock
        Examples:
            | staff with granted role           |
            | staff granted role school admin   |
            | staff granted role teacher        |
            | staff granted role hq staff       |
            | staff granted role centre manager |
            | staff granted role centre staff   |

    Scenario: "<staff with granted role>" send notification by role successfully
        Given current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        And a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff send notification
        Then returns "PermissionDenied" status code
        Examples:
            | staff with granted role         |
            | staff granted role teacher lead |
            | staff granted role centre lead  |

    @blocker
    Scenario Outline: create a notification and send it
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        And update user device token to an "valid" device token
        When current staff send notification
        Then notificationmgmt services must send notification to user
        And wait for FCM is sent to target user
        And recipient must receive the notification through FCM mock
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | all           | all          | all             | random       | random        | none               | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student, parent | random        | all          | all             | random       | random        | all                | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student, parent | all           | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student, parent | random        | random       | all             | random       | random        | none               | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student         | all           | all          | all             | random       | random        | none               | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student         | random        | all          | all             | random       | random        | all                | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student         | all           | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | parent          | all           | all          | all             | random       | random        | none               | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | parent          | random        | all          | all             | random       | random        | all                | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | parent          | all           | random       | all             | random       | random        | none               | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | parent          | random        | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student         | random        | random       | all             | random       | random        | all                | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student         | none          | none         | none            | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student, parent | none          | none         | none            | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |

    Scenario Outline: create a notification and send it after that check retention student/parent name on users_info_notifications
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        And update user device token to an "valid" device token
        When current staff send notification
        Then notificationmgmt services must send notification to user
        And username is saved follow by their notification
        And wait for FCM is sent to target user
        And recipient must receive the notification through FCM mock
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | all           | all          | all             | random       | random        | none               | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student         | all           | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | parent          | random        | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |

    Scenario Outline: create and send now notification fail
        Given current staff upsert "<status>" notification missing "<field>"
        Then returns "<code>" status code
        Examples:
            | status                    | field   | code            |
            | NOTIFICATION_STATUS_DRAFT | title   | InvalidArgument |
            | NOTIFICATION_STATUS_DRAFT | content | InvalidArgument |
