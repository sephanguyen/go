Feature: Assign assignment to topic

  Scenario Outline: Authentication for assign assignment to topic
    Given some assignments in db
    And a signed in "<role>"
    When assign assignment to topic
    Then returns "<status>" status code

    Examples: 
      | role           | status           |
      | school admin   | OK               |
      | student        | PermissionDenied |
      | parent         | PermissionDenied |
      | teacher        | PermissionDenied |
      | hq staff       | OK               |
      # | center lead    | PermissionDenied |
      # | center manager | PermissionDenied |
      # | center staff   | PermissionDenied |
