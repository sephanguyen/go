@quarantined
Feature: Withdraw order Frequency-base Package

  Scenario Outline: Withdraw order frequency-base package
    Given prepare data for withdraw order frequency-base package with "<valid data>"
    When "school admin" submit order
    Then withdraw order frequency-base package is created successfully
    And receives "OK" status code

    Examples:
      | valid data                                                 |
      | valid withdrawal request disabled prorating                |
      | valid withdrawal request with prorating and discount       |
      | valid graduate request with prorating and discount         |
      | empty billed at order with prorating and discount          |
      | empty upcoming billing with prorating and discount         |
      | empty billing items                                        |
      | duplicate products                                         |
