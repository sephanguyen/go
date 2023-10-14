@blocker
Feature: Update user group
    As a school admin
    I need to be able to update user group

    Scenario: update user group successfully
        Given a signed in "school admin"
        And a user group need to be updated
        When "school admin" update user group with valid payload
        Then "school admin" update user group successfully

    Scenario Outline: update user group successfully
        Given a signed in "school admin"
        And a user group was granted "<granted-role>" role need to be updated
        When "school admin" update user group with valid payload and grant "<update-granted-role>" role
        Then "school admin" update user group successfully
        And all users of that user group have "<legacy-usergroup>"

        Examples:
            | granted-role           | update-granted-role    | legacy-usergroup        |
            | Teacher                | School Admin           | USER_GROUP_SCHOOL_ADMIN |
            | School Admin           | Teacher                | USER_GROUP_TEACHER      |
            | Teacher                | School Admin, HQ Staff | USER_GROUP_SCHOOL_ADMIN |
            | School Admin           | Teacher, HQ Staff      | USER_GROUP_SCHOOL_ADMIN |
            | School Admin           | Teacher, Teacher Lead  | USER_GROUP_TEACHER      |
            | Teacher, Teacher Lead  | HQ Staff               | USER_GROUP_SCHOOL_ADMIN |
            | School Admin, HQ Staff | Teacher                | USER_GROUP_TEACHER      |

    Scenario Outline: user do not have permission to update user group
        Given a signed in "<role>"
        And a user group need to be updated
        When "<role>" update user group with valid payload
        Then "<role>" can not update user group and receive status code "<status code>" error

        Examples:
            | role    | status code      |
            | teacher | PermissionDenied |
            | student | PermissionDenied |
            | parent  | PermissionDenied |

    Scenario Outline: user update user group with invalid argument
        Given a signed in "school admin"
        And a user group need to be updated
        When "school admin" update user group with "<invalid type>" invalid argument
        Then "school admin" can not update user group and receive status code "<status code>" error

        Examples:
            | invalid type              | status code     |
            | user group id empty       | InvalidArgument |
            | user group is not existed | Internal        |
            | missing user group name   | InvalidArgument |
            | role id empty             | InvalidArgument |
            | location id empty         | InvalidArgument |
            | role is not existed       | InvalidArgument |
            | location is not existed   | InvalidArgument |
            | role missing location     | InvalidArgument |

    Scenario: user update user group successfully without argument
        Given a signed in "school admin"
        And a user group need to be updated
        When "school admin" update user group without argument "role with location"
        Then "school admin" update user group successfully
