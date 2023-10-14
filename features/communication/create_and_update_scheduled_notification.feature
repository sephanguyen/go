Feature: staff with granted role upsert scheduled notification

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    Scenario: "<staff with granted role>" upsert notification by role successfully
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_SCHEDULED" and important is "true"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Examples:
            | staff with granted role           |
            | staff granted role school admin   |
            | staff granted role teacher        |
            | staff granted role hq staff       |
            | staff granted role centre manager |
            | staff granted role centre staff   |

    Scenario: "<staff with granted role>" upsert notification by role <centre lead, teacher lead> will have permission denied
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_SCHEDULED" and important is "true"
        Then returns "PermissionDenied" status code
        Examples:
            | staff with granted role         |
            | staff granted role teacher lead |
            | staff granted role centre lead  |

    Scenario: staff with granted role upsert scheduled notification successfully
        When current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | all           | all          | random          | all          | all           | all                | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | all          | random          | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | all          | random          | all          | all           | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | random       | random          | all          | all           | all                | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | random       | random          | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | random       | random          | all          | all           | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | none         | random          | all          | all           | all                | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | none         | random          | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | none         | random          | all          | all           | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | all          | random          | all          | all           | all                | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | all          | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | all          | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | random       | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | random       | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | random       | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | none         | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | none         | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | random        | none         | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | all          | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | all          | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | all          | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | random       | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | random       | random          | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | random       | random          | all          | all           | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | none         | random          | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | all           | all          | random          | all          | all           | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | all           | all          | random          | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | all           | all          | random          | all          | all           | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | all           | random       | random          | all          | all           | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | all           | random       | random          | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | all           | random       | random          | all          | all           | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | all           | none         | random          | all          | all           | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | all           | none         | random          | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | all           | none         | random          | all          | all           | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | all          | random          | all          | all           | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | all          | random          | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | all          | random          | all          | all           | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | random       | random          | all          | all           | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | random       | random          | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | random       | random          | all          | all           | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | none         | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | none         | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | none         | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | none          | all          | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | none          | all          | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | none          | all          | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | none          | random       | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | none          | random       | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | none          | random       | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | none          | none         | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | all          | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | all          | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | all          | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | random       | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | random       | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | random       | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | none         | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | none         | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | all           | none         | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | random        | all          | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | random        | all          | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | random        | all          | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | random        | random       | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | random        | random       | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | random        | random       | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | random        | none         | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | random        | none         | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | random        | none         | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | none          | all          | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | none          | all          | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | none          | all          | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | none          | random       | random          | random       | random        | all                | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | none          | random       | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | none          | random       | random          | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | parent          | none          | none         | random          | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |

    Scenario: staff with granted role update notification successfully
        Given current staff upsert notification with valid filter for scheduled notification
        When current staff update notification with change "<field>"
        Then returns "OK" status code
        And update correctly corresponding field
        Examples:
            | field              |
            | content            |
            | title              |
            | user_groups        |
            | course_filter      |
            | grade_filter       |
            | individuals_filter |
            | status             |
            | scheduled_time     |
            | is_important       |
            | school_filter      |

    Scenario: scheduled notification has been sent, at this time user can no longer edit it
        Given current staff upsert notification with valid filter for scheduled notification
        When current staff send notification
        Then notificationmgmt services must send notification to user
        When current staff update notification with change "<field>"
        Then returns "InvalidArgument" status code
        And returns error message "the notification has been sent, you can no longer edit this notification"
        Examples:
            | field              |
            | content            |
            | title              |
            | user_groups        |
            | course_filter      |
            | grade_filter       |
            | individuals_filter |
            | status             |
            | scheduled_time     |
            | is_important       |

    Scenario: scheduled notification has been deleted, at this time user can no longer edit it
        Given current staff upsert notification with valid filter for scheduled notification
        When current staff discards notification
        Then notification is discarded
        When current staff update notification with change "<field>"
        Then returns "InvalidArgument" status code
        And returns error message "the notification has been deleted, you can no longer edit this notification"
        Examples:
            | field              |
            | content            |
            | title              |
            | user_groups        |
            | course_filter      |
            | grade_filter       |
            | individuals_filter |
            | status             |
            | scheduled_time     |
            | is_important       |

    Scenario: scheduled notification has invalid field scheduled_at
        When current staff upsert notification with invalid field scheduled_at which before current time
        Then returns "InvalidArgument" status code
        And returns error message "you cannot schedule at a time in the past"


    Scenario: scheduled notification has missed field scheduled_at
        When current staff upsert notification with missing field scheduled_at
        Then returns "InvalidArgument" status code
        And returns error message "request Notification.ScheduledAt time is empty"
