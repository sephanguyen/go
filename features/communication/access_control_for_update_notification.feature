Feature: staff with granted role update notification with security filter

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    # 1. The organization-level role updates their own notification
    @blocker
    Scenario: staff with organization level (school admin/hq staff) creates/updates a notification successfully (The organization-level role updates their own notification)
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "OK" status code
        And update correctly corresponding field
        Examples:
            | staff with granted role         | location_filter | location_filter_change |
            | staff granted role school admin | all             | 1                      |
            | staff granted role school admin | 1,2,3           | all                    |
            | staff granted role hq staff     | all             | 1,2                    |
            | staff granted role hq staff     | 1,2,3           | 1,2,3,4                |

    # 2. The organization-level role updates a notification that is created by another staff which the granted role is organization-level
    @blocker
    Scenario: staff with organization level (school admin/hq staff) creates/updates a notification successfully (The organization-level role updates a notification that is created by another staff which the granted role is organization-level)
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "OK" status code
        And update correctly corresponding field
        Examples:
            | staff with granted role         | location_filter | location_filter_change |
            | staff granted role school admin | all             | 1                      |
            | staff granted role school admin | 1,2,3           | all                    |
            | staff granted role hq staff     | all             | 1,2                    |
            | staff granted role hq staff     | 1,2,3           | 1,2,3,4                |

    @blocker
    # 3. The organization-level role updates notification that is created by another staff which the granted role is location-level
    Scenario: staff with organization level (school admin/hq staff) creates/updates a notification successfully (The organization-level role updates notification that is created by another staff which the granted role is location-level)
        Given a new "<staff with location-level granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Given a new "<staff with organization-level granted role>" and granted organization location logged in Back Office of a current organization
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "OK" status code
        And update correctly corresponding field
        Examples:
            | staff with location-level granted role | granted_location | staff with organization-level granted role | location_filter | location_filter_change |
            | staff granted role teacher             | 1,2,3            | staff granted role school admin            | all             | 1,2                    |
            | staff granted role teacher             | 1,2              | staff granted role school admin            | 1,2             | 1,2,3                  |
            | staff granted role teacher             | 1,2,3,4          | staff granted role hq staff                | all             | all                    |
            | staff granted role centre manager      | 1,2              | staff granted role hq staff                | 1,2             | default,1,2,3,4,5      |
            | staff granted role centre staff        | 1,2,3,4          | staff granted role hq staff                | 1,2,3,4         | 1,2                    |

    @blocker
    # 4. The location-level role updates their own notification with the valid location_filter
    Scenario: staff with location level (teacher/centre manager/centre staff) creates/updates a notification successfully (The location-level role updates their own notification with the valid location_filter)
        Given a new "<staff with location-level granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "OK" status code
        And update correctly corresponding field
        Examples:
            | staff with location-level granted role | granted_location | location_filter | location_filter_change |
            | staff granted role teacher             | 1,2,3            | all             | 1,2                    |
            | staff granted role teacher             | 1,2              | 1,2             | 1,2                    |
            | staff granted role teacher             | 1,2,3,4          | all             | all                    |
            | staff granted role teacher             | 1,2,3,4          | 1,2             | 1,2,3                  |
            | staff granted role centre manager      | 1,2              | 1,2             | 1                      |
            | staff granted role centre staff        | 1,2,3,4          | 1,2,3,4         | 1,2                    |

    # 5. The location-level role update their own notification with the invalid location_filter
    Scenario: staff with location level (teacher/centre manager/centre staff) creates/updates a notification violate access control (The location-level role update their own notification with the invalid location_filter)
        Given a new "<staff with location-level granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "Internal" status code
        Examples:
            | staff with location-level granted role | granted_location | location_filter | location_filter_change |
            | staff granted role teacher             | 1,2,3            | all             | default,1,2            |
            | staff granted role teacher             | 1,2              | 1,2             | 1,2,3                  |
            | staff granted role teacher             | 1,2,3,4          | all             | 4,5                    |
            | staff granted role teacher             | 1,2,3,4          | 1,2             | 3,4,5                  |
            | staff granted role centre manager      | 1,2              | 1,2             | 1,4,5                  |
            | staff granted role centre staff        | 1,2,3,4          | 1,2,3,4         | 1,2,5                  |

    # 6. The location-level role updates a notification that is created by another staff which the granted role is organization-level with the valid location_filter
    Scenario: staff with location level (teacher/centre manager/centre staff) creates/updates a notification violate access control (The location-level role updates a notification that is created by another staff which the granted role is organization-level with the valid location_filter)
        Given a new "<staff with organization-level granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Given a new "<staff with location-level granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "InvalidArgument" status code and error message have "PermissionDenied: Unauthorized to edit notification"
        Examples:
            | staff with organization-level granted role | location_filter | staff with location-level granted role | granted_location | location_filter_change |
            | staff granted role school admin            | 1,2             | staff granted role teacher             | 1,2              | 1                      |
            | staff granted role hq staff                | 1,2             | staff granted role centre manager      | 1,3,4,5          | 3,4                    |
            | staff granted role hq staff                | 3,4             | staff granted role centre staff        | 1,2,3,4          | 4                      |

    # 7. The location-level role updates a notification that is created by another staff which the granted role is organization-level with invalid location_filter
    Scenario: staff with location level (teacher/centre manager/centre staff) creates/updates a notification violate access control (The location-level role updates a notification that is created by another staff which the granted role is organization-level with invalid location_filter)
        Given a new "<staff with organization-level granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Given a new "<staff with location-level granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "InvalidArgument" status code and error message have "PermissionDenied: Unauthorized to edit notification"
        Examples:
            | staff with organization-level granted role | location_filter | staff with location-level granted role | granted_location | location_filter_change |
            | staff granted role school admin            | 1,2             | staff granted role teacher             | 1,2              | default,1              |
            | staff granted role hq staff                | 1,2,3           | staff granted role centre manager      | 3,4,5            | 3,5                    |
            | staff granted role hq staff                | 3,4             | staff granted role centre staff        | 1,2,3,4          | 1,4,5                  |

    # 8. The location-level role updates a notification that is created by another staff which the granted role is location-level with the valid location_filter
    Scenario: staff with location level (teacher/centre manager/centre staff) creates/updates a notification violate access control (The location-level role updates a notification that is created by another staff which the granted role is location-level with the valid location_filter)
        Given a new "<staff with location-level granted role 1>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Given a new "<staff with location-level granted role 2>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "InvalidArgument" status code and error message have "PermissionDenied: Unauthorized to edit notification"
        Examples:
            | staff with location-level granted role 1 | location_filter | staff with location-level granted role 2 | granted_location | location_filter_change |
            | staff granted role centre manager        | all             | staff granted role teacher               | 1,2,3            | 1,2                    |
            | staff granted role centre manager        | 1,2             | staff granted role teacher               | 1,2              | 1                      |
            | staff granted role teacher               | all             | staff granted role teacher               | 1,2,3,4          | all                    |
            | staff granted role teacher               | 3,4,5           | staff granted role centre manager        | 3,4,5            | 3,4                    |
            | staff granted role centre staff          | 3,4             | staff granted role centre staff          | 1,2,3,4          | 4                      |

    # 9. The location-level role updates a notification that is created by another staff which the granted role is location-level and with invalid location_filter
    Scenario: staff with location level (teacher/centre manager/centre staff) creates/updates a notification violate access control (The location-level role updates a notification that is created by another staff which the granted role is location-level and with invalid location_filter)
        Given a new "<staff with location-level granted role 1>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Given a new "<staff with location-level granted role 2>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "InvalidArgument" status code and error message have "PermissionDenied: Unauthorized to edit notification"
        Examples:
            | staff with location-level granted role 1 | location_filter | staff with location-level granted role 2 | granted_location | location_filter_change |
            | staff granted role teacher               | all             | staff granted role teacher               | 1,2,3            | 1,2,5                  |
            | staff granted role teacher               | 1,2             | staff granted role teacher               | 1,2              | default,1              |
            | staff granted role centre staff          | all             | staff granted role teacher               | 1,2,3,4          | 4,5                    |
            | staff granted role hq staff              | 3,4,5           | staff granted role centre manager        | 3,4,5            | 1,2,3                  |
            | staff granted role centre manager        | 3,4             | staff granted role centre staff          | 1,2,3,4          | 1,4,5                  |
