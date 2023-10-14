Feature: List Topics By Study Plan

  Background: 
    Given a validated book with chapters and topics

  Scenario Outline: Authentication for user try to list topics by study plan
    Given a signed in "<role>"
    When user list topics by study plan
    Then returns "<status>" status code

    Examples: 
      | role           | status           |
      | school admin   | OK               |
      | student        | PermissionDenied |
      | parent         | PermissionDenied |
      | teacher        | OK               |
      | hq staff       | OK               |
      # | center lead    | OK               |
      # | center manager | OK               |
      # | center staff   | OK               |

  Scenario: user try to list topics by study plan
    Given a signed in "teacher"
    When user list topics by study plan
    Then returns "OK" status code
    And verify topic data after list topics by study plan
