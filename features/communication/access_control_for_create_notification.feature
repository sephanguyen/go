Feature: staff with granted role create notification with security filter

    Background:
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "random" students with "1" parents info for each student
        And school admin creates "random" courses
        And school admin add packages data of those courses for each student

    # 1. The organization-level role creates a notification with the valid location_filter
    @blocker
    Scenario: staff with organization level (school admin/hq staff) creates a notification successfully (The organization-level role creates a notification with the valid location_filter)
        Given a new "<staff with granted role>" and granted organization location logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Examples:
            | staff with granted role         | location_filter |
            | staff granted role school admin | all             |
            | staff granted role school admin | 1,2,3           |
            | staff granted role hq staff     | all             |
            | staff granted role hq staff     | 1,2,3           |

    # 2. The location-level role creates a notification with the valid location_filter
    @blocker
    Scenario: staff with location level (teacher/centre manager/centre staff) is granted descendant locations creates a notification successfully (The location-level role creates a notification with the valid location_filter)
        Given a new "<staff with granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "OK" status code
        And notificationmgmt services must store the notification with correctly info
        Examples:
            | staff with granted role           | granted_location | location_filter |
            | staff granted role teacher        | 1,2,3            | all             |
            | staff granted role teacher        | 1,2,3            | 1,2             |
            | staff granted role teacher        | 1,2,3            | 1,2,3           |
            | staff granted role centre manager | 1,2,3,4          | 1,2,3           |
            | staff granted role centre staff   | 1,2,3,4          | 1,2,3           |

    # 3. The location-level role creates a notification with the invalid location_filter (location out of range user's granted location)
    @blocker
    Scenario: staff with location level (teacher/centre manager/centre staff) is granted descendant locations creates a notification violate access control (The location-level role creates a notification with the invalid location_filter (location out of range user's granted location))
        Given a new "<staff with granted role>" and granted "<granted_location>" descendant locations logged in Back Office of a current organization
        When current staff upsert notification to "student, parent" and "random" course and "random" grade and "<location_filter>" location and "random" class and "random" school and "random" individuals and "random" scheduled time with "NOTIFICATION_STATUS_DRAFT" and important is "false"
        Then returns "Internal" status code
        Examples:
            | staff with granted role           | granted_location | location_filter |
            | staff granted role teacher        | 1,2,3            | default,1,2,3   |
            | staff granted role teacher        | 1,2,3            | 1,2,3,4         |
            | staff granted role teacher        | 1,2,3            | 1,2,3,4,5       |
            | staff granted role teacher        | 1,2,3            | 1,2,3,4         |
            | staff granted role centre manager | 1,2,3            | 1,2,3,4         |
            | staff granted role centre staff   | 1,2,3            | 1,2,3,4         |
