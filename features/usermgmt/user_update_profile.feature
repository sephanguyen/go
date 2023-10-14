Feature: Update User Profile
  As a internal system user
  I need to be able to update my profile

  Scenario Outline: User update own profile
    Given a signed in "<role>"
    When the signed in user update profile
    Then user update profile successfully

    Examples:
      | role    |
      | student |
      | parent  |

  Scenario Outline: User cannot update profile of another user
    Given a signed in "<role>"
    When the signed in user update another user profile
    Then user cannot update user profile

    Examples:
      | role    |
      | student |
      | parent  |

  Scenario Outline: User cannot update profile without mandatory field
    Given a signed in "<role>"
    When the signed in user update user profile without mandatory "<field>" field
    Then user cannot update user profile

    Examples:
      | role    | field |
      | student | name  |
      | parent  | name  |
