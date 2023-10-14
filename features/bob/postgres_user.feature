Feature: Get postgres users info

  Scenario: Get postgres users info with invalid base64 format key
    Given a request to get postgres users info
    And an invalid base64 format postgres user info key
    When call get postgres users info API
    Then returns "InvalidArgument" status code

  Scenario: Get postgres users info with invalid RSA format key
    Given a request to get postgres users info
    And an invalid RSA format postgres user info key
    When call get postgres users info API
    Then returns "InvalidArgument" status code

  Scenario: Get postgres users info with invalid key
    Given a request to get postgres users info
    And an invalid postgres user info key
    When call get postgres users info API
    Then returns "PermissionDenied" status code

  Scenario: Get postgres users info with valid key
    Given a request to get postgres users info
    And a valid postgres user info key
    When call get postgres users info API
    Then returns "OK" status code
    And postgres users info data must contains "bob"
    And postgres users info data must contains "tom"
    And postgres users info data must contains "eureka"
    And postgres users info data must contains "fatima"
    And postgres users info data must contains "postgres"
    And postgres users info data must contains "zeus"
    And postgres users info data must contains "draft"
    And postgres users info data must contains "shamir"
    And postgres users info data must contains "hasura"
    And postgres users info data must contains "kafka_connector"

  Scenario: Get postgres privilege with invalid base64 format key
    Given a request to get postgres privilege
    And an invalid base64 format postgres user info key
    When call get postgres privilege by API
    Then returns "InvalidArgument" status code

  Scenario: Get postgres privilege with invalid RSA format key
    Given a request to get postgres privilege
    And an invalid RSA format postgres user info key
    When call get postgres privilege by API
    Then returns "InvalidArgument" status code

  Scenario: Get postgres privilege with invalid key
    Given a request to get postgres privilege
    And an invalid postgres user info key
    When call get postgres privilege by API
    Then returns "PermissionDenied" status code

  Scenario: Get postgres privilege with valid key
    Given a request to get postgres privilege
    And a valid postgres user info key
    When call get postgres privilege by API
    Then returns "OK" status code

  Scenario: Get postgres privilege with valid key
    Given a request to get postgres privilege
    And a valid postgres user info key
    When call get postgres privilege by API
    Then returns "OK" status code
    And postgres privilege info data must contains "bob"
    And postgres privilege info data must contains "hasura"
    And postgres privilege info data must contains "kafka_connector"