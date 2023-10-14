Feature: Order base student course subscription

  Scenario Outline: Create order frequency-base package success
    Given prepare data for create order frequency-base package
    When "school admin" submit order
    Then order frequency-base package is created successfully
    And receives "OK" status code
    And package upserted to student package table

  Scenario Outline: Update order recurring package for student subscription success
    Given prepare data for create order recurring package
    And "school admin" submit order
    And receives "OK" status code
    And prepare data for update order recurring package
    When "school admin" submit order
    Then receives "OK" status code
    And update recurring package success
    And package upserted to student package table

  Scenario Outline: Cancel order recurring package for student subscription success
    Given prepare data for create order recurring package
    And "school admin" submit order
    And receives "OK" status code
    And prepare data for cancel order recurring package
    When "school admin" submit order
    Then receives "OK" status code
    And update recurring package success
    And package upserted to student package table

  Scenario Outline: Withdrawal order for student subscription success
    Given prepare data for withdraw order recurring package
    When "school admin" submit order
    Then withdraw order frequency-base package is created successfully
    And receives "OK" status code
    And package upserted to student package table

  Scenario Outline: Graduation order for student subscription success
    Given prepare data for graduate order recurring package
    When "school admin" submit order
    Then withdraw order frequency-base package is created successfully
    And receives "OK" status code
    And package upserted to student package table