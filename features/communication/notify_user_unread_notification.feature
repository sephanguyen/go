Feature: notify user unread notification after we create and send notification, some users read notification but some users do not. We want to notify user who has not read notification

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    Scenario: "<staff with granted role>" notify unread user notification by role successfully
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        And current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        When current staff send notification
        Then notificationmgmt services must send notification to user
        And some users read notification
        When current staff notifies notification to unread users
        Then returns "OK" status code
        Examples:
            | staff with granted role           |
            | staff granted role school admin   |
            | staff granted role teacher        |
            | staff granted role hq staff       |
            | staff granted role centre manager |
            | staff granted role centre staff   |

    Scenario: "<staff with granted role>" notify unread user notification by role successfully
        Given current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        When current staff send notification
        Then notificationmgmt services must send notification to user
        And some users read notification
        And a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff notifies notification to unread users
        Then returns "PermissionDenied" status code
        Examples:
            | staff with granted role         |
            | staff granted role teacher lead |
            | staff granted role centre lead  |

    Scenario: notify user who unread notification successfully
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        When current staff send notification
        Then notificationmgmt services must send notification to user
        And some users read notification
        And update user device token to an "valid" device token
        When current staff notifies notification to unread users
        Then returns "OK" status code
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | none          | none         | none            | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student, parent | random        | random       | default         | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student         | random        | random       | default         | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | parent          | random        | random       | default         | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |

    Scenario: notify user who unread notification failed fcm batch error
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        When current staff send notification
        Then notificationmgmt services must send notification to user
        And some users read notification
        And update user device token to an "invalid" device token
        When current staff notifies notification to unread users
        Then returns "OK" status code
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | none          | none         | none            | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student, parent | random        | random       | default         | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student         | random        | random       | default         | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | parent          | random        | random       | default         | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |

    Scenario: notify user who unread notification failed fcm
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        When current staff send notification
        Then notificationmgmt services must send notification to user
        And update user device token to an "unexpected" device token
        When current staff notifies notification to unread users
        Then returns "Internal" status code
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | random        | random       | default         | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | student         | random        | random       | default         | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
            | parent          | random        | random       | default         | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |
