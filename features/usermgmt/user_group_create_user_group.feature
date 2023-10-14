@blocker
Feature: Create user group
  As a school admin
  I need to be able to create user group

  Scenario Outline: create user group successfully
    Given a signed in "<role>"
    When signed in user create "<validity>" user group
    Then returns "OK" status code
    And user group is created successfully
    And user group after creating must be existed in database correctly

    Examples:
      | role         | validity |
      | school admin | valid    |

  Scenario Outline: Can not create user group if user is not school admin
    Given a signed in "<role>"
    When signed in user create "<validity>" user group
    Then returns "<status code>" status code

    Examples:
      | role                 | validity | status code      |
      | unauthenticated      | valid    | Unauthenticated  |
      | student              | valid    | PermissionDenied |
      | parent               | valid    | PermissionDenied |
      | teacher              | valid    | PermissionDenied |
      # | organization manager | valid    | PermissionDenied |

  Scenario Outline: Can not create user group with invalid params
    Given a signed in "<role>"
    When signed in user create "<validity>" user group
    Then returns "<status code>" status code

    Examples:
      | role         | validity            | status code      |
      | school admin | missing name        | InvalidArgument  |
      | school admin | missing role_ids    | InvalidArgument  |
      | school admin | invalid location_id | InvalidArgument  |
      | school admin | invalid role_id     | InvalidArgument  |
