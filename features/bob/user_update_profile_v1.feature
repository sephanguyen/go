Feature: Update User Profile

  Scenario Outline: User update profile of another user
    Given a signed in user "<role>" with school: <school for login>
    And a profile of "other" user with usergroup: "<user group>", name: "<name>", phone: "<phone>", email: "<email>", school: <school>
    When user update profile
    Then profile of "other" user must be updated
    And event "EvtUserInfo" must be published to "UserDeviceToken.Updated" channel

    Examples:
      | role         | school for login | user group              | name           | phone  | email                      | school |
      | school admin | 3                | USER_GROUP_TEACHER      | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | school admin | 3                | USER_GROUP_TEACHER      | update-user-%d | +849%d | update-user-%d@example.com | 3      |

  Scenario Outline: User cannot update profile of another user without proper permission
    Given a signed in user "<role>" with school: <school for login>
    And a profile of "other" user with usergroup: "<user group>", name: "<name>", phone: "<phone>", email: "<email>", school: <school>
    When user update profile
    Then returns "<error>" status code

    Examples:
      | role         | school for login | user group              | name           | phone  | email                      | school | error            |
      | school admin | 3                | USER_GROUP_SCHOOL_ADMIN | update-user-%d | +849%d | update-user-%d@example.com | 3      | PermissionDenied |
      | school admin | 3                | USER_GROUP_TEACHER      | update-user-%d | +849%d | update-user-%d@example.com | 2      | PermissionDenied |
      | student      | 3                | USER_GROUP_STUDENT      | update-user-%d | +849%d | update-user-%d@example.com | 2      | PermissionDenied |

  Scenario Outline: User update own profile
    Given a signed in user "<role>" with school: <school for login>
    And a profile of "own" user with usergroup: "<user group>", name: "<name>", phone: "<phone>", email: "<email>", school: <school>
    When user update profile
    Then profile of "own" user must be updated
    And event "EvtUserInfo" must be published to "UserDeviceToken.Updated" channel

    Examples:
      | role         | school for login | user group              | name           | phone  | email                      | school |
      | school admin | 3                | USER_GROUP_SCHOOL_ADMIN | update-user-%d | +849%d |                            | 3      |
      | school admin | 3                | USER_GROUP_SCHOOL_ADMIN | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | teacher      | 3                | USER_GROUP_TEACHER      | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | student      | 3                | USER_GROUP_STUDENT      | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | student      | 3                | USER_GROUP_STUDENT      | update-user-%d |        | update-user-%d@example.com | 3      |
      | parent       | 3                | USER_GROUP_PARENT       | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | parent       | 3                | USER_GROUP_PARENT       | update-user-%d |        | update-user-%d@example.com | 3      |

 