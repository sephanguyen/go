Feature: staff with granted role upsert notification

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student

    Scenario: "<staff with granted role>" upsert notification by role successfully
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
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
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "PermissionDenied" status code
        Examples:
            | staff with granted role         |
            | staff granted role teacher lead |
            | staff granted role centre lead  |

    @blocker
    Scenario: staff with granted role upsert notification successfully
        When current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | all           | all          | all             | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | all           | random       | all             | all          | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | random        | all          | all             | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | none         | all             | all          | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | all           | none         | 1,2,3           | none         | all           | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | all          | 1,2             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student         | random        | random       | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | parent          | random        | random       | 1,2,3,4         | none         | none          | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | random       | 1,2,3           | random       | none          | random             | random           | NOTIFICATION_STATUS_DRAFT     | false        |
            | student         | all           | all          | all             | random       | all           | random             | random           | NOTIFICATION_STATUS_DRAFT     | true         |
            | student         | all           | all          | all             | none         | all           | random             | none             | NOTIFICATION_STATUS_DRAFT     | true         |

    Scenario: staff with granted role upsert notification fail
        When current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        Then returns "InvalidArgument" status code
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | random        | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_SCHEDULED | false        |
            | none            | random        | random       | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | none            | random        | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student         | random        | random       | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SENT      | true         |
            | student         | random        | random       | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_NONE      | true         |
            | none            | all           | all          | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT     | true         |
            | none            | none          | none         | all             | none         | random        | none               | none             | NOTIFICATION_STATUS_DRAFT     | false        |

    @blocker
    Scenario: staff with granted role update notification successfully
        Given current staff upsert notification with valid filter
        When current staff update notification with change "<field>"
        Then returns "OK" status code
        And update correctly corresponding field
        Examples:
            | field                         |
            | content                       |
            | title                         |
            | user_groups                   |
            | course_filter                 |
            | grade_filter                  |
            | class_filter                  |
            | location_filter               |
            | individuals_filter            |
            | status                        |
            | scheduled_time                |
            | is_important                  |
            | excluded_generic_receiver_ids |
            | school_filter                 |

    Scenario: staff with granted role update notification successfully
        Given current staff upsert notification with valid filter
        When current staff update notification with change "<field>"
        Then returns "OK" status code
        And update correctly corresponding field
        When current staff update notification filter with change selection all
        Then notificationmgmt services must store the notification with correctly info
        Examples:
            | field           |
            | course_filter   |
            | class_filter    |
            | location_filter |

    Scenario: staff with granted role upsert notification with retention individual name successfully
        When current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        And individual name is saved successfully
        When current staff upsert notification again with new individual targets
        Then returns "OK" status code
        And individual name is saved successfully
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | none          | none         | all             | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT | true         |
            | student         | none          | none         | all             | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | parent          | none          | none         | all             | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT | true         |

    Scenario: staff with granted role create notification with generic_receiver_id and then update it
        When current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        When current staff upsert notification again with new individual targets
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | none          | none         | all             | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
