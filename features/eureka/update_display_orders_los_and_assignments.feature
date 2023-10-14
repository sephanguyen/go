Feature: Update display orders Los And Assignments
  Background: valid book in db
    Given "school admin" logins "CMS"
    And valid book in bob
    And user create los and assignments

  Scenario Outline: Authentication for update display orders los and assignments
    Given "<role>" logins "CMS"
    When user update display orders for los and assignments
    Then returns "<status>" status code

    Examples: 
      | role           | status           |
      | school admin   | OK               |
      | student        | PermissionDenied |
      | parent         | PermissionDenied |
      | teacher        | OK               |
      | hq staff       | OK               |
      # | center lead    | PermissionDenied |
      # | center manager | PermissionDenied |
      # | center staff   | PermissionDenied |

    Scenario: Update display orders los and assignments
      When user update display orders for los and assignments
      Then returns "OK" status code
      And display order of los and assignments must be updated
