@quarantined
Feature: As a user, I want to get access token in salesforce system

  Scenario: Staff exchange salesforce token
    Given a user signed in as "<signed-in user>" in "manabie" organization
    When user exchanges salesforce token
    Then user exchanges salesforce token successfully

    Examples:
      | signed-in user                  |
      | staff granted role school admin |
      | staff granted role hq staff     |

  Scenario: student, parent can't exchange salesforce token
    Given a user signed in as "<signed-in user>" in "manabie" organization
    When user exchanges salesforce token
    Then user can not exchanges salesforce token with status code "<code>"

    Examples:
      | signed-in user | code             |
      | student        | PermissionDenied |
      | parent         | PermissionDenied |
