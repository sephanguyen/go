Feature: User join student conversation

  Background: Student has teacher in conversation
    Given resource path of school "Manabie" is applied
    And a student conversation with 2 teacher

  @blocker
  Scenario Outline: teacher and school admin join conversation successfully
    Given a valid "<user>" token
    When "<user>" join conversations
    Then returns "OK" status code
    And "<user>" must be member of conversations
    And system must send "joined" conversation message

    Examples:
      | user         |
      | teacher      |
      | school admin |

  Scenario Outline: user do not have permission to join conversation
    Given a valid "<user>" token
    When "<user>" join conversations
    Then returns "PermissionDenied" status code

    Examples:
      | user    |
      | student |
      | parent  |
