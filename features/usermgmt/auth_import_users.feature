Feature: Import users to a tenant

  Scenario: Import user from tenant to another tenant
    Given users in tenant 1
    When admin import users from tenant 1 to tenant 2
    Then users in tenant 1 still have valid info
    And users in tenant 2 has corresponding info

  Scenario: Import user from firebase auth to identity platform tenant
    Given users in firebase auth
    When admin import users from firebase auth to tenant in identity platform
    Then users in firebase auth still still have valid info
    And users in tenant has corresponding info

  @blocker
  Scenario: Create auth profiles by import operation successfully
    When system create auth profiles with valid info
    Then auth profiles are created successfully and users can use them to login in to system

  @blocker
  Scenario Outline: Failed to create auth profiles by import operation
    When system create auth profiles "<invalid case>"
    Then system failed to create auth profiles and users can not use them to login in to system

    Examples:
      | invalid case                                    |
      | but there is a profile has empty user id        |
      | but there is a profile has invalid email format |

  @blocker
  Scenario: Update auth profiles by import operation
    Given existing auth profiles in system
    When system update existing auth profiles with valid info
    Then auth profiles are updated successfully and users can use them to login in to system

  @blocker
  Scenario Outline: Update auth profiles by import operation, email or password is changed
    Given existing auth profiles in system
    And user already logged in with existing auth profile
    When system update existing auth profiles with valid info "<profile condition>"
    Then auth profiles are updated successfully and user can use them login in, if already logged in, user have to login in again

    Examples:
      | profile condition            |
      | but only email is changed    |
      | but only password is changed |

  @blocker
  Scenario: Update auth profiles by import operation but both email and password are not changed
    Given existing auth profiles in system
    And user already logged in with existing auth profile
    When system update existing auth profiles with valid info "but email and password are not changed"
    Then auth profiles are updated successfully but user doesn't need to login in again, also they still can login in again with old profile

  @blocker
  Scenario Outline: Failed to update auth profiles by import operation
    Given existing auth profiles in system
    When system update existing auth profiles "<invalid case>"
    Then system failed to update auth profiles and users can not use them to login in to system

    Examples:
      | invalid case                                    |
      | but there is a profile has empty user id        |
      | but there is a profile has invalid email format |
