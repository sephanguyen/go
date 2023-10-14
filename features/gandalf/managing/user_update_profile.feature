Feature: Update User Profile

  Scenario Outline: users update his profile
    Given a signed in "<role>" with school: <school for login>
    And a request update "his own" profile with user group: "<user group>", name: "<name>", phone: "<phone>", email: "<email>", school: <school>
    When user ask Bob to do update
    Then Bob must update "his own" profile
    And Bob must publish event to user_device_token channel
    And Tom must record new user_device_tokens with updated name info

    Examples:
      | role         | school for login | user group              | name           | phone  | email                      | school |
      | school admin | 3                | USER_GROUP_SCHOOL_ADMIN | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | admin        | 3                | USER_GROUP_ADMIN        | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | teacher      | 3                | USER_GROUP_TEACHER      | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | student      | 3                | USER_GROUP_STUDENT      | update-user-%d | +849%d | update-user-%d@example.com | 3      |

  Scenario Outline: users update another profile
    Given a signed in "<role>" with school: <school for login>
    And a request update "other" profile with user group: "<user group>", name: "<name>", phone: "<phone>", email: "<email>", school: <school>
    When user ask Bob to do update
    Then Bob must update "other" profile
    And Bob must publish event to user_device_token channel
    And Tom must record new user_device_tokens with updated name info

    Examples:
      | role         | school for login | user group              | name           | phone  | email                      | school |
      | admin        | 3                | USER_GROUP_SCHOOL_ADMIN | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | admin        | 3                | USER_GROUP_TEACHER      | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | admin        | 3                | USER_GROUP_STUDENT      | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | school admin | 3                | USER_GROUP_TEACHER      | update-user-%d | +849%d | update-user-%d@example.com | 3      |
      | school admin | 3                | USER_GROUP_TEACHER      | update-user-%d | +849%d | update-user-%d@example.com | 3      |




