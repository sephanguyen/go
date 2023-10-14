Feature: user view recipient list when composing notification
    @blocker
    Scenario: user create a notification with target group filters, and view recipient list in popup
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert notification to "<user_groups>" and "<course_filter>" course and "<grade_filter>" grade and "<location_filter>" location and "<class_filter>" class and "<school_filter>" school and "<individuals_filter>" individuals and "<scheduled_status>" scheduled time with "<status>" and important is "<is important>"
        Then returns "OK" status code
        And current staff view recipient list in detail page of draft or scheduled notification
        When current staff upsert and send notification
        Then recipients must be same as data from view recipient list in detail page
        Examples:
            | user_groups     | course_filter | grade_filter | location_filter | class_filter | school_filter | individuals_filter | scheduled_status | status                        | is important |
            | student, parent | all           | all          | all             | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student, parent | none          | none         | all             | none         | none          | random             | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | random       | all             | none         | none          | none               | random           | NOTIFICATION_STATUS_SCHEDULED | false        |
            | parent          | random        | random       | all             | none         | none          | none               | random           | NOTIFICATION_STATUS_SCHEDULED | true         |
            | student         | random        | random       | all             | none         | none          | random             | none             | NOTIFICATION_STATUS_DRAFT     | false        |

    Scenario: user view recipient list and use paging, case none individual
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "<num_student>" students with "1" parents info for each student
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert notification to "student, parent" and "all" course and "all" grade and "default" location and "all" class and "none" school and "none" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then current staff get the audience list in detail page with "<page_no>" and "<limit>" and see results display "<count_result>" rows
        Examples:
            | num_student | page_no | limit | count_result |
            | 5           | 1       | 10    | 10           |
            | 5           | 2       | 4     | 4            |
            | 4           | 3       | 3     | 2            |

    Scenario: user view recipient list and use paging, case none group audience
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        Given current staff upsert notification to "student, parent" and "none" course and "none" grade and "default" location and "none" class and "none" school and "2" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then current staff get the audience list in detail page with "<page_no>" and "<limit>" and see results display "<count_result>" rows
        Examples:
            | page_no | limit | count_result |
            | 1       | 10    | 4            |

    Scenario: user view recipient list with parent have multiple students
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "<num_student>" students with the same parent
        And school admin creates "2" students with "1" parents info for each student
        And school admin creates "1" students with "3" parents info for each student
        And school admin creates "random" courses with "1" classes for each course
        And school admin add packages data of those courses for each student
        Given current staff upsert notification to "student, parent" and "all" course and "all" grade and "default" location and "all" class and "none" school and "none" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then current staff get the audience list in detail page with "<page_no>" and "<limit>" and see results display "<count_result>" rows
        Examples:
            | num_student | page_no | limit | count_result |
            | 5           | 1       | 20    | 18           |
            | 5           | 2       | 4     | 4            |
            | 2           | 4       | 3     | 3            |