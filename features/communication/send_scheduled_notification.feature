Feature: staff with granted role send notification

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    Scenario: "<staff with granted role>" send notification by role successfully
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_SCHEDULED" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
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

    Scenario Outline: staff with granted role send notification using all filter
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        And update user device token to an "valid" device token
        When current staff send notification
        Then notificationmgmt services must send notification to user
        And wait for FCM is sent to target user
        And recipient must receive the notification through FCM mock
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | all           | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | all          | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | random       | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | random       | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | none         | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | none         | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | all          | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | none          | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | none          | all          | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student         | all           | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student         | all           | all          | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student         | all           | random       | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student         | all           | random       | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student         | all           | none         | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student         | all           | none         | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | all          | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | none          | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | none          | all          | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | all          | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | random       | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | random       | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | none         | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | none         | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | random        | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | random        | all          | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | none          | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | none          | all          | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | none         | none            | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | none          | none         | none            | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | none          | none         | none            | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |

    Scenario Outline: staff with granted send notification. After that user cannot send notification
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        And update user device token to an "valid" device token
        When current staff send notification
        Then notificationmgmt services must send notification to user
        When current staff send notification again
        Then returns "InvalidArgument" status code
        And returns error message "the notification has been sent"
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | random        | random       | default         | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |

    Scenario Outline: staff with granted discard notification. After that user cannot send notification
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        When current staff discards notification
        Then notification is discarded
        When current staff send notification again
        Then returns "InvalidArgument" status code
        And returns error message "the notification has been deleted, you can no longer send this notification"
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | random        | random       | default         | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
