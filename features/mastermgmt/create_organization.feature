@quarantined
Feature: Create Organization
  Scenario: Recieving origanization
    Given a generate school
    And a signed in as "organization manager"
    When user create new organization
    Then location type and location default created successfully
    And Mastermgmt must push msg "UpsertLocation" subject "SyncLocation.Upserted" to nats
    And Mastermgmt must push msg "UpsertLocationType" subject "SyncLocationType.Upserted" to nats

  Scenario Outline: Create new organization
    Given new organization data
    When "<signed-in user>" create new organization
    Then new organization were created successfully

    Examples:
      | signed-in user       |
      | organization manager |


  Scenario Outline: user create organization success without fields
    Given organization data has empty or invalid "<fields>"
    When "<signed-in user>" create new organization
    Then new organization were created successfully

    Examples:
      | signed-in user       | fields         |
      | organization manager | logo url       |
      | organization manager | country code   |

  Scenario Outline: user create a organization fail with some fields are missing
    Given organization data has empty or invalid "<fields>"
    When "<signed-in user>" create new organization
    Then returns "<msg>" status code

    Examples:
      | signed-in user       | fields            | msg             |
      | organization manager | tenantID          | InvalidArgument |
      | organization manager | organization name | InvalidArgument |


  Scenario Outline: user create a organization fail with invalid domain_name
    Given organization data has invalid domain name "<domain_name>"
    When "<signed-in user>" create new organization
    Then returns "<msg>" status code

    Examples:
      | signed-in user       | domain_name       | msg     |
      | organization manager | -test-domain-name | Unknown |
      | organization manager | test-DOMAIN       | Unknown |
      | organization manager | test-domain-name  | OK      |
      | organization manager | test-domain-9999  | OK      |


  Scenario Outline: School admin, student, parent, teacher don't have permission to create organization
    Given new organization data
    When "<signed-in user>" user can not create organization
    Then returns "<msg>" status code

    Examples:
      | signed-in user  | msg              |
      | admin           | PermissionDenied |
      | teacher         | PermissionDenied |
      | parent          | PermissionDenied |
      | student         | PermissionDenied |
      | school admin    | PermissionDenied |
      | unauthenticated | Unauthenticated  |


  Scenario Outline: Organization manager have permission to create organization
    Given new organization data
    When "<signed-in user>" create new organization
    Then returns "<msg>" status code
    Examples:
      | signed-in user       | msg |
      | organization manager | OK  |
    