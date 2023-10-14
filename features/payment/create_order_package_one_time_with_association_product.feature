Feature: Create order associated product of package list in student billing

  Scenario: Create order one time package with association product success
    Given prepare data for create order one time package with association product
    When "school admin" submit order
    Then receives "OK" status code
    And package upserted to student package table
    And "school admin" get list order package with association product
    And check response data of successfully
    #And an event must be published to signal student course sync

  Scenario: Create order one time package with association product and duplicated product success
    Given prepare data for create order one time package with association product and duplicated product
    When "school admin" submit order
    Then receives "OK" status code
    And package upserted to student package table
    And "school admin" get list order package with association product
    And check response data of successfully
    #And an event must be published to signal student course sync

  Scenario: Create order one time package with association recurring product success
    Given prepare data for create order one time package with association recurring product
    When "school admin" submit order
    Then receives "OK" status code
    And package upserted to student package table
    And "school admin" get list order package with association product
    And check response data of successfully
    #And an event must be published to signal student course sync