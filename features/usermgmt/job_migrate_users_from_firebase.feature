Feature: Run migration job to migrate users from firebase auth

  Scenario: Migrate users from firebase auth to identity platform
    Given users in our system have been imported to firebase auth
    When system run job to migrate users from firebase auth
    Then info of users in firebase auth is still valid
    And info of users in tenant of identity platform has corresponding info