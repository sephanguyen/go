Feature: user view recipient list when composing notification
    @blocker
    Scenario: user create a notification with target group filters, and view recipient list in popup
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff compose notification with "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        And current staff view recipient list popup
        When current staff upsert and send notification
        Then recipients must be same as data from view recipient list popup
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | all           | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | all           | random       | all             | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | none         | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | all          | all             | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student         | random        | random       | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | parent          | random        | random       | all             | none         | none          | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | random       | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_DRAFT     | false        |
            | student         | all           | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_DRAFT     | true         |
            | student         | all           | all          | all             | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT     | true         |

    Scenario: user view recipient list and use search feature
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff compose notification with "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        And current staff view recipient list popup
        Then current staft search for a keyword and see correct result
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | all           | all          | all             | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | all           | random       | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student, parent | all           | none         | all             | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | student         | random        | random       | all             | none         | none          | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | parent          | random        | random       | all             | random       | random        | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | random       | all             | none         | none          | random             | random           | NOTIFICATION_STATUS_DRAFT     | false        |
            | student         | all           | all          | all             | random       | random        | random             | random           | NOTIFICATION_STATUS_DRAFT     | true         |
            | student         | all           | all          | all             | random       | random        | random             | none             | NOTIFICATION_STATUS_DRAFT     | true         |

    Scenario: user view recipient list and use paging
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "<num_student>" students with "1" parents info for each student
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        When current staff compose notification with "student, parent" and "all" course and "all" grade and "default" location and "all" class and "none" school and "none" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then current staff get the audience list with "<page_no>" and "<limit>" and see results display "<count_result>" rows
        Examples:
            | num_student | page_no | limit | count_result |
            | 5           | 1       | 10    | 10           |
            | 5           | 2       | 4     | 4            |
            | 4           | 3       | 3     | 2            |

    Scenario: user use RetrieveGroupAudience API to check for excluded receiver id
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff compose notification with "student, parent" and "all" course and "all" grade and "none" location and "all" class and "none" school and "none" individuals and "none" scheduled time with "NOTIFICATION_STATUS_SCHEDULED" and important is "false"
        And current staff view recipient list popup
        And school admin excluded some recipients from the list
        When excluded recipients no longer available to view
        Then api RetrieveGroupAudience return empty result
