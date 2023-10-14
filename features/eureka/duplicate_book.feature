Feature: Duplicate book
  All content in books need to be copy
  Background:
    Given a signed in "school admin"
    And a valid book in db

  Scenario Outline: Authentication for copy all content in book
    Given a signed in "<role>"
    When user send duplicate book request
    Then returns "<status>" status code

    Examples:
      | role           | status           |
      | school admin   | OK               |
      | student        | PermissionDenied |
      | parent         | PermissionDenied |
      | teacher        | OK               |
      | hq staff       | OK               |
      | center lead    | PermissionDenied |
      | center manager | PermissionDenied |
      | center staff   | PermissionDenied |

  Scenario: Copy all content in book
    When user send duplicate book request
    Then returns "OK" status code
    And eureka must return copied topics

