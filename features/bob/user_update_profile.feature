Feature: Update User Profile

  Scenario Outline: users update another profile with wrong cases
    Given a signed in "<role>" with school: <school for login>
    And user updated "other" profile with user group: "<user group>", name: "<name>", phone: "<phone>", email: "<email>", school: <school>
    When user ask Bob to do update
    Then returns "<error>" status code

    Examples:
      | role         | school for login | user group              | name           | email                      | school | error            |
      | school admin | 3                |                         | update-user-%d | update-user-%d@example.com | 3      | InvalidArgument  |
      | school admin | 3                | USER_GROUP_ADMIN        | update-user-%d | update-user-%d@example.com | 3      | PermissionDenied |
      | school admin | 3                | USER_GROUP_ADMIN        | update-user-%d | update-user-%d@example.com | 3      | PermissionDenied |
      | school admin | 3                | USER_GROUP_SCHOOL_ADMIN | update-user-%d | update-user-%d@example.com | 3      | PermissionDenied |
      | school admin | 3                | USER_GROUP_TEACHER      | update-user-%d | update-user-%d@example.com | 2      | PermissionDenied |
      | student      | 3                | USER_GROUP_STUDENT      | update-user-%d | update-user-%d@example.com | 2      | PermissionDenied |

  Scenario Outline: users update hist own profile with wrong cases
    Given a signed in "<role>" with school: <school for login>
    And user updated "his own" profile with user group: "<user group>", name: "<name>", phone: "<phone>", email: "<email>", school: <school>
    When user ask Bob to do update
    Then returns "<error>" status code

    Examples:
      | role         | school for login | user group         | name           | email                      | school | error           |
      | school admin | 3                | USER_GROUP_TEACHER |                | update-user-%d@example.com | 3      | InvalidArgument |
      | school admin | 3                |                    | update-user-%d | update-user-%d@example.com | 3      | InvalidArgument |
      | school admin | 3                |                    |                | update-user-%d@example.com | 3      | InvalidArgument |
      | school admin | 3                |                    |                |                            | 3      | InvalidArgument |
      | teacher      | 3                | USER_GROUP_TEACHER |                | update-user-%d@example.com | 3      | InvalidArgument |
      | teacher      | 3                |                    | update-user-%d | update-user-%d@example.com | 3      | InvalidArgument |
      | teacher      | 3                |                    |                | update-user-%d@example.com | 3      | InvalidArgument |
      | teacher      | 3                |                    |                |                            | 3      | InvalidArgument |
      | student      | 3                | USER_GROUP_STUDENT |                | update-user-%d@example.com | 3      | InvalidArgument |
      | student      | 3                |                    | update-user-%d | update-user-%d@example.com | 3      | InvalidArgument |
      | student      | 3                |                    |                | update-user-%d@example.com | 3      | InvalidArgument |
      | student      | 3                |                    |                |                            | 3      | InvalidArgument |


  Scenario Outline: users update his profile
    Given a signed in "<role>" with school: <school for login>
    And user updated "his own" profile with user group: "<user group>", name: "<name>", phone: "<phone>", email: "<email>", school: <school>
    When user ask Bob to do update
    Then Bob must update "his own" profile
    And Bob must publish event to user_device_token channel

    Examples:
      | role         | school for login | user group              | name           | phone  | email                      | school |
      | school admin | 3                | USER_GROUP_SCHOOL_ADMIN | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | school admin | 3                | USER_GROUP_SCHOOL_ADMIN | update-user-%d |        | update-user-%d@example.com | 3      |
      | school admin | 3                | USER_GROUP_TEACHER      | update-user-%d | +849%d |                            | 3      |
      | school admin | 3                | USER_GROUP_TEACHER      | update-user-%d |        | update-user-%d@example.com | 3      |
      | teacher      | 3                | USER_GROUP_TEACHER      | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | teacher      | 3                | USER_GROUP_TEACHER      | update-user-%d |        | update-user-%d@example.com | 3      |
      | student      | 3                | USER_GROUP_STUDENT      | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | student      | 3                | USER_GROUP_STUDENT      | update-user-%d |        | update-user-%d@example.com | 3      |

