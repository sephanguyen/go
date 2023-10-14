Feature: staff with granted role discard draft, scheduled notification
    user want to discard draft, scheduled notification

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    Scenario: "<staff with granted role>" discard notification by role successfully
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        And current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        When current staff discards notification
        Then notification is discarded
        Examples:
            | staff with granted role           |
            | staff granted role school admin   |
            | staff granted role teacher        |
            | staff granted role hq staff       |
            | staff granted role centre manager |
            | staff granted role centre staff   |

    Scenario:  "<staff with granted role>" discard notification by role <centre lead, teacher lead> will have permission denied
        Given current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        And a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff discards notification
        Then returns "PermissionDenied" status code
        Examples:
            | staff with granted role         |
            | staff granted role teacher lead |
            | staff granted role centre lead  |

    @blocker
    Scenario: discard draft/scheduled notification
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        When current staff discards notification
        Then notification is discarded
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | random        | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT     | true         |
            | student, parent | random        | random       | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | none         | all             | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT     | true         |
            | student, parent | none          | none         | all             | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |


    Scenario Outline: cannot discard sent notification
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        When current staff send notification
        Then notificationmgmt services must send notification to user
        When current staff discards notification
        Then returns "InvalidArgument" status code
        And returns error message "the notification has been sent, you can no longer discard this notification"
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | random        | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | true         |


    Scenario Outline: staff with granted role discard notification, some other user cannot discard notification again
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        When current staff discards notification
        Then notification is discarded
        When current staff discards notification
        Then returns "InvalidArgument" status code
        And returns error message "the notification has been deleted"
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | random        | random       | random          | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | true         |

    Scenario Outline: staff with granted role discard a notification which not their owner
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        And a new "staff granted role school admin" and granted organization location logged in Back Office of a current organization
        When current staff discards notification
        Then returns "OK" status code
        And notification is discarded
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | random        | random       | random          | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | true         |
