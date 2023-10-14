Feature: exchange token

  As a user, I need to exchange token before calling other API

  @quarantined
  Scenario: exchange token with valid authentication token
    Given an other student profile in DB
    And a valid authentication token with ID already exist in DB
    When a user exchange token
    Then our system need to do return valid token
  @blocker
  Scenario: school admin exchange token
    Given an school admin profile in DB
    When a user exchange token
    Then our system need to do return valid "school admin" exchanged token

  @blocker
  Scenario: teacher exchange token
    Given a signed in teacher
    When a user exchange token
    Then our system need to do return valid "teacher" exchanged token

  @quarantined
  Scenario Outline: user in tenant exchange token
    Given "<user>" in "<auth platform>" logged in
    When "<user>" uses id token to exchanges token with our system
    Then "<user>" receives valid exchanged token

    Examples:
      | user                | auth platform |
      | legacy student      | identity      |
      | legacy student      | firebase      |
      | legacy parent       | identity      |
      | legacy parent       | firebase      |
      | legacy teacher      | identity      |
      | legacy teacher      | firebase      |
      | legacy school admin | identity      |
      | legacy school admin | firebase      |
      | student             | identity      |
      | student             | firebase      |
      | parent              | identity      |
      | parent              | firebase      |
      | teacher             | identity      |
      | teacher             | firebase      |
      | school admin        | identity      |
      | school admin        | firebase      |

  @blocker
  Scenario Outline: user in tenant exchange token
    Given "<user>" in keycloak logged in
    When "<user>" uses id token to exchanges token with our system
    Then "<user>" receives valid exchanged token

    Examples:
      | user                   |
      | legacy student         |
      | legacy teacher         |
      | legacy school admin    |
      | student                |
      | teacher                |
      | school admin           |
      | student with kids type |
      | student with a+ type   |

  @blocker
  Scenario Outline: check default values for auth info
    When system init default values for auth info in "<env>"
    Then the initialized values must be valid

    Examples:
      | env   |
      | local |
      | stag  |
      | uat   |
      | prod  |
