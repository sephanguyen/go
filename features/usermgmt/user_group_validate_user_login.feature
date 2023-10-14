@blocker
Feature: Validate user login with specifying platform
  As a logined user
  I need to check me can access to the platform

  # @feature_flag:User_Auth_AllowAllRolesToLoginTeacherWeb
  Scenario Outline: use defaut group in user to validate login
    Given a signed in "<role>"
    # And "disable" Unleash feature with feature name "User_Auth_AllowAllRolesToLoginTeacherWeb"
    When check this user is able to access "<platform-name>" platform
    Then returns "OK" status code
    And user is "<status>" to login this platform

    Examples:
      | role                              | platform-name       | status |
      | staff granted role school admin   | PLATFORM_BACKOFFICE | able   |
      | staff granted role hq staff       | PLATFORM_BACKOFFICE | able   |
      | staff granted role centre lead    | PLATFORM_BACKOFFICE | able   |
      | staff granted role centre manager | PLATFORM_BACKOFFICE | able   |
      | staff granted role centre staff   | PLATFORM_BACKOFFICE | able   |
      | staff granted role teacher lead   | PLATFORM_BACKOFFICE | able   |
      | staff granted role teacher        | PLATFORM_BACKOFFICE | able   |
      | school admin                      | PLATFORM_BACKOFFICE | able   |
      | teacher                           | PLATFORM_BACKOFFICE | able   |
      | student                           | PLATFORM_BACKOFFICE | unable |
      | parent                            | PLATFORM_BACKOFFICE | unable |
      # | staff granted role school admin   | PLATFORM_TEACHER  | unable |
      # | staff granted role hq staff       | PLATFORM_TEACHER  | unable |
      # | staff granted role centre lead    | PLATFORM_TEACHER  | unable |
      # | staff granted role centre manager | PLATFORM_TEACHER  | unable |
      # | staff granted role centre staff   | PLATFORM_TEACHER  | unable |
      # | staff granted role teacher lead   | PLATFORM_TEACHER  | unable |
      | staff granted role teacher        | PLATFORM_TEACHER    | able |
      # | school admin                      | PLATFORM_TEACHER    | unable |
      | teacher                           | PLATFORM_TEACHER    | able |
      | student                           | PLATFORM_TEACHER    | unable |
      | parent                            | PLATFORM_TEACHER    | unable |
      | staff granted role school admin   | PLATFORM_LEARNER    | unable |
      | staff granted role hq staff       | PLATFORM_LEARNER    | unable |
      | staff granted role centre lead    | PLATFORM_LEARNER    | unable |
      | staff granted role centre manager | PLATFORM_LEARNER    | unable |
      | staff granted role centre staff   | PLATFORM_LEARNER    | unable |
      | staff granted role teacher lead   | PLATFORM_LEARNER    | unable |
      | staff granted role teacher        | PLATFORM_LEARNER    | unable |
      | school admin                      | PLATFORM_LEARNER    | unable |
      | teacher                           | PLATFORM_LEARNER    | unable |
      | student                           | PLATFORM_LEARNER    | able   |
      | parent                            | PLATFORM_LEARNER    | able   |

  @quarantined
  Scenario Outline: user who have not been assigned user_group is unable to access any platform
    Given a signed in "<role>"
    When check this user is able to access "<platform-name>" platform
    Then user is "unable" to login this platform

    Examples:
      | role         | platform-name       |
      | school admin | PLATFORM_BACKOFFICE |
      | teacher      | PLATFORM_BACKOFFICE |
      | student      | PLATFORM_BACKOFFICE |
      | parent       | PLATFORM_BACKOFFICE |
      | school admin | PLATFORM_TEACHER    |
      | teacher      | PLATFORM_TEACHER    |
      | student      | PLATFORM_TEACHER    |
      | parent       | PLATFORM_TEACHER    |
      | school admin | PLATFORM_LEARNER    |
      | teacher      | PLATFORM_LEARNER    |
      | student      | PLATFORM_LEARNER    |
      | parent       | PLATFORM_LEARNER    |

  # @feature_flag:User_Auth_AllowAllRolesToLoginTeacherWeb
  Scenario Outline: use defaut group in user to validate login
    Given a signed in "<role>"
    # And "able" Unleash feature with feature name "User_Auth_AllowAllRolesToLoginTeacherWeb"
    When check this user is able to access "<platform-name>" platform
    Then returns "OK" status code
    And user is "<status>" to login this platform

    Examples:
      | role                              | platform-name       | status |
      | staff granted role school admin   | PLATFORM_BACKOFFICE | able   |
      | staff granted role hq staff       | PLATFORM_BACKOFFICE | able   |
      | staff granted role centre lead    | PLATFORM_BACKOFFICE | able   |
      | staff granted role centre manager | PLATFORM_BACKOFFICE | able   |
      | staff granted role centre staff   | PLATFORM_BACKOFFICE | able   |
      | staff granted role teacher lead   | PLATFORM_BACKOFFICE | able   |
      | staff granted role teacher        | PLATFORM_BACKOFFICE | able   |
      | school admin                      | PLATFORM_BACKOFFICE | able   |
      | teacher                           | PLATFORM_BACKOFFICE | able   |
      | student                           | PLATFORM_BACKOFFICE | unable |
      | parent                            | PLATFORM_BACKOFFICE | unable |
      | staff granted role school admin   | PLATFORM_TEACHER    | able   |
      | staff granted role hq staff       | PLATFORM_TEACHER    | able   |
      | staff granted role centre lead    | PLATFORM_TEACHER    | able   |
      | staff granted role centre manager | PLATFORM_TEACHER    | able   |
      | staff granted role centre staff   | PLATFORM_TEACHER    | able   |
      | staff granted role teacher lead   | PLATFORM_TEACHER    | able   |
      | staff granted role teacher        | PLATFORM_TEACHER    | able   |
      | school admin                      | PLATFORM_TEACHER    | able   |
      | teacher                           | PLATFORM_TEACHER    | able   |
      | student                           | PLATFORM_TEACHER    | unable |
      | parent                            | PLATFORM_TEACHER    | unable |
      | staff granted role school admin   | PLATFORM_LEARNER    | unable |
      | staff granted role hq staff       | PLATFORM_LEARNER    | unable |
      | staff granted role centre lead    | PLATFORM_LEARNER    | unable |
      | staff granted role centre manager | PLATFORM_LEARNER    | unable |
      | staff granted role centre staff   | PLATFORM_LEARNER    | unable |
      | staff granted role teacher lead   | PLATFORM_LEARNER    | unable |
      | staff granted role teacher        | PLATFORM_LEARNER    | unable |
      | school admin                      | PLATFORM_LEARNER    | unable |
      | teacher                           | PLATFORM_LEARNER    | unable |
      | student                           | PLATFORM_LEARNER    | able   |
      | parent                            | PLATFORM_LEARNER    | able   |
