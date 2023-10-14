@quarantined
Feature: Get user profile

  Scenario: user try to get profile with wrong cases
    Given signed as "unauthenticated" account
    When user get profile
    Then returns "Unauthenticated" status code

  Scenario Outline: user try to get profile
    Given signed as "<signed as>" account
    When user get profile
    Then returns "OK" status code
    And yasuo must return user profile

    Examples:
      | signed as    |
      | school admin |
      | teacher      |

  Scenario Outline: user try to get profile with user group data
    Given signed as "<signed as>" account have user group
    When  user get profile
    Then returns "OK" status code
    And yasuo must return user profile

    Examples:
      | signed as    |
      | teacher      |