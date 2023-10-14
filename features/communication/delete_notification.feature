Feature: staff with granted role delete draft, scheduled notification
    user want to delete draft, scheduled notification

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    Scenario: "<staff with granted role>" delete notification by role successfully
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        And current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        When current staff send notification
        And current staff deletes notification
        Then notification is deleted
        And recipient must not retrieve notification
        Examples:
            | staff with granted role           |
            | staff granted role school admin   |
            | staff granted role teacher        |
            | staff granted role hq staff       |
            | staff granted role centre manager |
            | staff granted role centre staff   |

    Scenario:  "<staff with granted role>" delete notification by role <centre lead, teacher lead> will have permission denied
        Given current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        And a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff deletes notification
        Then returns "PermissionDenied" status code
        Examples:
            | staff with granted role         |
            | staff granted role teacher lead |
            | staff granted role centre lead  |

    Scenario: delete draft/scheduled notification
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        When current staff deletes notification
        Then notification is deleted
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | random        | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT     | false        |
            | student, parent | random        | random       | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | none          | none         | all             | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT     | false        |
            | student, parent | none          | none         | all             | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |

    Scenario Outline: delete sent notification
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        When current staff send notification
        Then notificationmgmt services must send notification to user
        When current staff deletes notification
        Then returns "OK" status code
        And notification is deleted
        And recipient must not retrieve notification
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | random        | random       | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | false        |

    Scenario Outline: staff with granted role delete notification, some other user cannot delete notification again
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        When current staff deletes notification
        Then notification is deleted
        When current staff deletes notification
        Then returns "InvalidArgument" status code
        And returns error message "the notification has been deleted"
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | random        | random       | random          | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | true         |

    Scenario Outline: staff with granted role delete a notification which not their owner
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        And a new "staff granted role school admin" and granted organization location logged in Back Office of a current organization
        When current staff deletes notification
        Then returns "OK" status code
        And notification is deleted
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                    | is important |
            | student, parent | random        | random       | random          | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT | true         |

    Scenario: staff delete notification with questionnaire
        Given current staff create a questionnaire with resubmit allowed "<resubmit_allowed>", questions "<questions>" respectively
        And current staff upsert notification to "student, parent" and "random" course and "random" grade and "none" location and "none" class and "none" school and "random" individuals and "none" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        When current staff send notification
        And current staff deletes notification
        Then notification is deleted
        And recipient must not retrieve notification
        Examples:
            | resubmit_allowed | questions                                                                           |
            | false            | 1.multiple_choice, 2.free_text.required, 3.check_box.required                       |
            | true             | 1.free_text, 2.free_text.required, 3.check_box.required, 4.multiple_choice.required |