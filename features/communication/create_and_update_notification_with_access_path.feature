Feature: staff with granted role upsert notification with access path

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    Scenario: staff with organization level (school admin/hq staff) create a notification
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        And the notification access path has been store correctly with "<location_access_path>" locations
        Examples:
            | staff with granted role         | location_filter | location_access_path |
            | staff granted role school admin | all             | default              |
            | staff granted role school admin | 1,2             | 1,2                  |
            | staff granted role school admin | 1,2,3           | 1,2,3                |
            | staff granted role hq staff     | all             | default              |
            | staff granted role hq staff     | 1,2             | 1,2                  |

    Scenario: staff with organization level (school admin/hq staff) update a notification
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        And the notification access path has been store correctly with "<location_access_path_before>" locations
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "OK" status code
        And update correctly corresponding field
        And the notification access path has been store correctly with "<location_access_path_after>" locations
        Examples:
            | staff with granted role         | location_filter | location_access_path_before | location_filter_change | location_access_path_after |
            | staff granted role school admin | all             | default                     | 1,2                    | 1,2                        |
            | staff granted role school admin | 1,2             | 1,2                         | 1,2,3,4                | 1,2,3,4                    |
            | staff granted role school admin | 1,2,3           | 1,2,3                       | all                    | default                    |
            | staff granted role school admin | 1,2,3           | 1,2,3                       | 1                      | 1                          |
            | staff granted role hq staff     | 1,2,3           | 1,2,3                       | 1                      | 1                          |
            | staff granted role hq staff     | 1,2             | 1,2                         | all                    | default                    |

    Scenario: staff with location level (teacher/centre manager/centre staff) is granted descendant locations create a notification
        Given a new "<staff with granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        And the notification access path has been store correctly with "<location_access_path>" locations
        Examples:
            | staff with granted role           | location_filter | granted_location | location_access_path |
            | staff granted role teacher        | all             | 1,2,3            | 1,2,3                |
            | staff granted role teacher        | 1,2             | 1,2,3            | 1,2                  |
            | staff granted role teacher        | 1,2,3           | 1,2,3            | 1,2,3                |
            | staff granted role centre manager | 1,2,3           | 1,2,3,4          | 1,2,3                |
            | staff granted role centre staff   | 1,2,3           | 1,2,3,4          | 1,2,3                |

    Scenario: staff with location level (teacher/centre manager/centre staff) is granted descendant locations update a notification
        Given a new "<staff with granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        And the notification access path has been store correctly with "<location_access_path_before>" locations
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "OK" status code
        And update correctly corresponding field
        And the notification access path has been store correctly with "<location_access_path_after>" locations
        Examples:
            | staff with granted role           | location_filter | granted_location | location_access_path_before | location_filter_change | location_access_path_after |
            | staff granted role teacher        | all             | 1,2,3            | 1,2,3                       | 1,2                    | 1,2                        |
            | staff granted role teacher        | 1,2             | 1,2,3            | 1,2                         | all                    | 1,2,3                      |
            | staff granted role teacher        | 1,2,3           | 1,2,3            | 1,2,3                       | all                    | 1,2,3                      |
            | staff granted role teacher        | 1,2,3           | 1,2,3,4          | 1,2,3                       | 1                      | 1                          |
            | staff granted role centre manager | 1,2,3           | 1,2,3            | 1,2,3                       | 1                      | 1                          |
            | staff granted role centre staff   | 1,2,3,4         | 1,2,3,4          | 1,2,3,4                     | 1,2,3                  | 1,2,3                      |

    Scenario: staff with location level (teacher/centre manager/centre staff) is granted descendant locations create draft/scheduled notification and then update granted locations
        Given a new "<staff with granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_SCHEDULED" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        And the notification access path has been store correctly with "<location_access_path_before>" locations
        When admin update staff granted locations to "<granted_location_change>"
        And current staff upsert with "<location_filter_change>" locations and send notification
        Then returns "OK" status code
        And the notification access path has been store correctly with "<location_access_path_after>" locations
        Examples:
            | staff with granted role           | location_filter | granted_location | location_access_path_before | granted_location_change | location_filter_change | location_access_path_after |
            | staff granted role teacher        | all             | 1,2,3            | 1,2,3                       | 1,2,3,4                 | all                    | 1,2,3,4                    |
            | staff granted role teacher        | 1,2             | 1,2,3            | 1,2                         | 2,3,4                   | 2                      | 2                          |
            | staff granted role centre manager | all             | 1,2,3            | 1,2,3                       | 1,2,3,4                 | 1                      | 1                          |
            | staff granted role centre staff   | all             | 1,2,3            | 1,2,3                       | 1,2,3,4                 | 1,2,3,4                | 1,2,3,4                    |
            | staff granted role teacher        | 1               | 1,2              | 1                           | 3,4,5                   | 4,5                    | 4,5                        |

    Scenario: staff with location level (teacher/centre manager/centre staff) updates their own notification after it was updated by organization level role
        Given a new "<staff with location-level granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        When "school admin" update the notification with location filter change to "4,5"
        Then "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "OK" status code
        And the notification access path has been store correctly with "<location_filter_check>" locations
        Examples:
            | staff with location-level granted role | granted_location | location_filter | location_filter_change | location_filter_check |
            | staff granted role teacher             | 1,2,3            | all             | 1,2                    | 1,2                   |
            | staff granted role teacher             | 1,2              | 1,2             | 1,2                    | 1,2                   |
            | staff granted role teacher             | 1,2,3,4          | all             | all                    | 1,2,3,4               |
            | staff granted role teacher             | 1,2,3,4          | 1,2             | 1,2,3                  | 1,2,3                 |
            | staff granted role centre manager      | 1,2              | 1,2             | 1                      | 1                     |
            | staff granted role centre staff        | 1,2,3,4          | 1,2,3,4         | 1,2                    | 1,2                   |
