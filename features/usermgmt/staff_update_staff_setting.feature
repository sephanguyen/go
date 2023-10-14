Feature: Update Staff Setting

  Background: Prepare staff
    Given prepare students data

  Scenario Outline: user update staff setting
    Given a signed in "<role>"
    And a staff config with staff id: "<staff_id>"
    When user update staff config
    Then returns "<msg>" status code

    Examples:
      | role                            | staff_id | msg              |
      | teacher                         | exist    | PermissionDenied |
      | student                         | exist    | PermissionDenied |
      | parent                          | exist    | PermissionDenied |
      | staff granted role teacher      | exist    | PermissionDenied |
      | staff granted role school admin | empty    | InvalidArgument  |
      | staff granted role school admin | exist    | OK               |

