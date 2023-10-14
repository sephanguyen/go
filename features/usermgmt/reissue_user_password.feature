@blocker
Feature: Reissue user password
  In order to reissue user's password
  As a school admin
  I need to perform a update to change user's password

  Scenario Outline: School admin, or school staff reissues user's password
    Given a signed in "staff granted role school admin"
    And the signed in user create "<account type>" user
    When "<role>" reissues user's password
    Then receives "OK" status code
    And user can sign in with the new password

    Examples:
      | role                            | account type               | code |
      | staff granted role school admin | student                    | OK   |
      | staff granted role hq staff     | parent                     | OK   |
      | staff granted role centre lead  | staff granted role teacher | OK   |

  Scenario Outline: Owner profile can reissues user's password
    Given a signed in "staff granted role school admin"
    And the signed in user create "<account type>" user
    When the owner reissues user's password
    Then receives "OK" status code
    And user can sign in with the new password

    Examples:
      | account type |
      | student      |
      | parent       |

  Scenario Outline: Teacher, student, parent don't have permission to reissue another user's password
    Given a signed in "staff granted role school admin"
    And the signed in user create "<account type>" user
    When "<role>" reissues user's password
    Then receives "<code>" status code

    Examples:
      | role            | account type | code             |
      | teacher         | student      | PermissionDenied |
      | parent          | student      | PermissionDenied |
      | student         | student      | PermissionDenied |
      | unauthenticated | student      | Unauthenticated  |

  Scenario Outline: Admin cannot reissue password of non-existing user
    Given a signed in "staff granted role school admin"
    And the signed in user create "<account type>" user
    When "<role>" reissues user's password with non-existing user
    Then receives "Internal" status code

    Examples:
      | role         | account type               |
      | school admin | student                    |
      | school admin | parent                     |
      | school admin | staff granted role teacher |
  
  Scenario Outline: Admin cannot reissue user's password with missing parameters
    Given a signed in "staff granted role school admin"
    And the signed in user create "<account type>" user
    When "<role>" reissues user's password when missing "<field name>" field
    Then receives "InvalidArgument" status code

    Examples:
      | role         | account type               |
      | school admin | student                    |
      | school admin | parent                     |
      | school admin | staff granted role teacher |
