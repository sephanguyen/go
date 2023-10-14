Feature: staff with granted role deletes notification with security filter

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    # 1. The organization-level role deletes their own notification
    Scenario: staff with organization level (school admin/hq staff) deletes a notification successfully (The organization-level role deletes their own notification)
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        When "current" staff discards notification
        Then returns "OK" status code
        Examples:
            | staff with granted role         | location_filter |
            | staff granted role school admin | all             |
            | staff granted role school admin | 1,2,3           |
            | staff granted role hq staff     | all             |
            | staff granted role hq staff     | 1,2,3           |

    # 2. The organization-level role deletes a notification that is created by another staff which the granted role is organization-level
    Scenario: staff with organization level (school admin/hq staff) deletes a notification successfully (The organization-level role deletes a notification that is created by another staff which the granted role is organization-level)
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When "current" staff discards notification
        Then returns "OK" status code
        Examples:
            | staff with granted role         | location_filter |
            | staff granted role school admin | all             |
            | staff granted role school admin | 1,2,3           |
            | staff granted role hq staff     | all             |
            | staff granted role hq staff     | 1,2,3           |

    # 3. The organization-level role get a notification that is created by another staff which the granted role is location-level
    Scenario: staff with organization level (school admin/hq staff) deletes a notification successfully (The organization-level role get a notification that is created by another staff which the granted role is location-level)
        Given a new "<staff with location-level granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Given a new "<staff with organization-level granted role>" and granted organization location logged in Back Office of a current organization
        When "current" staff discards notification
        Then returns "OK" status code
        Examples:
            | staff with location-level granted role | granted_location | staff with organization-level granted role | location_filter |
            | staff granted role teacher             | 1,2,3            | staff granted role school admin            | all             |
            | staff granted role teacher             | 1,2              | staff granted role school admin            | 1,2             |
            | staff granted role teacher             | 1,2,3,4          | staff granted role hq staff                | all             |
            | staff granted role centre manager      | 1,2              | staff granted role hq staff                | 1,2             |
            | staff granted role centre staff        | 1,2,3,4          | staff granted role hq staff                | 1,2,3,4         |

    # 4. The location-level role deletes their own notification
    Scenario: staff with location level (teacher/centre manager/centre staff) deletes a notification successfully (The location-level role deletes their own notification)
        Given a new "<staff with location-level granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        When "current" staff discards notification
        Then returns "OK" status code
        Examples:
            | staff with location-level granted role | granted_location | location_filter |
            | staff granted role teacher             | 1,2,3            | all             |
            | staff granted role teacher             | 1,2              | 1,2             |
            | staff granted role teacher             | 1,2,3,4          | all             |
            | staff granted role teacher             | 1,2,3,4          | 1,2             |
            | staff granted role centre manager      | 1,2              | 1,2             |
            | staff granted role centre staff        | 1,2,3,4          | 1,2,3,4         |

    # 5. The location-level role deletes their own notification, in case the location-level role creates a notification, after that the organization-level role updates this notification with another location that is different/the same from the location-level granted location
    Scenario: staff with location level (teacher/centre manager/centre staff) deletes a notification successfully (The location-level role deletes their own notification, in case the location-level role creates a notification, after that the organization-level role updates this notification with another location that is different/the same from the location-level granted location)
        Given a new "<staff with location-level granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Given a new "<staff with organization-level granted role>" and granted organization location logged in Back Office of a current organization
        When "current staff" update the notification with location filter change to "<location_filter_change>"
        Then returns "OK" status code
        And update correctly corresponding field
        When "previous" staff discards notification
        Then returns "OK" status code
        Examples:
            | staff with location-level granted role | granted_location | staff with organization-level granted role | location_filter | location_filter_change |
            | staff granted role teacher             | 1,2,3            | staff granted role school admin            | all             | 4,5                    |
            | staff granted role teacher             | 1,2              | staff granted role school admin            | 1,2             | 3                      |
            | staff granted role teacher             | 1,2,3,4          | staff granted role hq staff                | all             | 5                      |
            | staff granted role centre manager      | 1,2              | staff granted role hq staff                | 1,2             | default,1,2,3,4,5      |
            | staff granted role centre staff        | 1,2,3,4          | staff granted role hq staff                | 1,2,3,4         | 1,2                    |

    # 6. The location-level role deletes a notification in their granted location that is created by another staff which the granted role is organization-level
    Scenario: staff with location level (teacher/centre manager/centre staff) deletes a notification not successfully (The location-level role deletes a notification in their granted location that is created by another staff which the granted role is organization-level)
        Given a new "<staff with organization-level granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Given a new "<staff with location-level granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When "current" staff discards notification
        Then returns "InvalidArgument" status code and error message have "PermissionDenied: Unauthorized to discard notification"
        Examples:
            | staff with organization-level granted role | location_filter | staff with location-level granted role | granted_location |
            | staff granted role school admin            | 1,2             | staff granted role teacher             | 1,2              |
            | staff granted role hq staff                | 1,2             | staff granted role centre manager      | 1,3,4,5          |
            | staff granted role hq staff                | 3,4             | staff granted role centre staff        | 3,4,5            |

    # 7. The location-level role deletes a notifications in their granted location that is created by another staff which the granted role is location-level
    Scenario: staff with location level (teacher/centre manager/centre staff) deletes a notification not successfully (The location-level role deletes a notifications in their granted location that is created by another staff which the granted role is location-level)
        Given a new "<staff with location-level granted role 1>" and granted "<granted_location 1>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Given a new "<staff with location-level granted role 2>" and granted "<granted_location 2>" descendant locations logged in Back Office of a current organization
        When "current" staff discards notification
        Then returns "InvalidArgument" status code and error message have "PermissionDenied: Unauthorized to discard notification"
        Examples:
            | staff with location-level granted role 1 | granted_location 1 | location_filter | staff with location-level granted role 2 | granted_location 2 |
            | staff granted role teacher               | 3,4,5              | 3,4             | staff granted role teacher               | 1,2,3              |
            | staff granted role teacher               | 1,2                | 1,2             | staff granted role teacher               | 1,2                |
            | staff granted role centre staff          | 1,2,5              | all             | staff granted role teacher               | 2,3,4              |
            | staff granted role centre staff          | 2,3,5              | 3,5             | staff granted role centre manager        | 3,4,5              |
            | staff granted role centre manager        | 1,2,3,4            | 3,4             | staff granted role centre staff          | 1,2,3,4,5          |
