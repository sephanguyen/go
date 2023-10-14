Feature: Create timesheet

  Background:
    Given have timesheet configuration is on

  Scenario Outline: Staff create a timesheet for themselves
    Given "<signed-in user>" signin system
    And new timesheet data for current staff
    When user creates a new timesheet
    Then returns "<resp status-code>" status code
    And the timesheet is created "<timesheet-created>"

    Examples:
      | signed-in user                    | resp status-code | timesheet-created |
      | staff granted role school admin   | OK               | true              |
      | staff granted role teacher        | OK               | false             |

  Scenario Outline: Staff create a timesheet for other staff
    Given "<signed-in user>" signin system
    And new timesheet data for other staff
    When user creates a new timesheet
    Then returns "<resp status-code>" status code
    And the timesheet is created "<timesheet-created>"

    Examples:
      | signed-in user                    | resp status-code | timesheet-created |
      | staff granted role school admin   | OK               | true              |
      | staff granted role teacher        | PermissionDenied | false             |

  Scenario Outline: Staff timesheet with already existing timesheet
    Given "<signed-in user>" signin system
    And new timesheet data for existing timesheet
    When user creates a new timesheet
    Then returns "<resp status-code>" status code

    Examples:
      | signed-in user                    | resp status-code |
      | staff granted role school admin   | AlreadyExists    |

  Scenario Outline: Invalid user create a timesheet for themselves
    Given "<invalid user>" signin system
    When user creates a new timesheet
    Then returns "<resp status-code>" status code

    Examples:
      | invalid user    | resp status-code |
      | unauthenticated | Unauthenticated  |
      | parent          | PermissionDenied |

  Scenario Outline: Invalid user create a timesheet other user
    Given "<invalid user>" signin system
    When user creates a new timesheet
    Then returns "<resp status-code>" status code

    Examples:
      | invalid user    | resp status-code |
      | unauthenticated | Unauthenticated  |
      | parent          | PermissionDenied |
