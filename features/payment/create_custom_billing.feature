Feature: Create custom billing

  Scenario: CreateCustomBilling success
    Given prepare data for creating custom billing
    When "school admin" submit custom billing request
    Then custom billing is created successfully
    And receives "OK" status code

  Scenario Outline: CreateCustomBilling failure
    Given request for create custom billing with "<invalid data>"
    When "school admin" submit custom billing request
    Then receives "<status code>" status code
    And receives "<error message>" error message for create custom billing with "<invalid data>"

    Examples:
      | invalid data                              | status code          | error message                                                              |
      | not-exist student                         | Internal             | Error when checking student id: row.Scan: no rows in result set            |
      | not-exist location                        | Internal             | Error when checking location id: row.Scan: no rows in result set           |
      | invalid tax category                      | FailedPrecondition   | Product with name %v changed tax category from %v to %v                    |
      | invalid tax percentage                    | FailedPrecondition   | Product with name %v change tax percentage from %v to %v                   |
      | invalid tax amount                        | FailedPrecondition   | Incorrect tax amount actual = %v vs expected = %v                          |
      | missing location                          | FailedPrecondition   | Missing mandatory data: location                                           |
      | missing name                              | FailedPrecondition   | Missing mandatory data: custom billing item name                           |

  Scenario: CreateCustomBilling with account category success
    Given prepare data for creating custom billing with account category
    When "school admin" submit custom billing request
    Then custom billing is created successfully
    And receives "OK" status code
