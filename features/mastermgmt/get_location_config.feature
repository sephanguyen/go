Feature: Get all configurations

  Background: Given condition
    Given location configurations value existed on DB

  Scenario Outline: Get all configurations
    Given "school admin" signin system
    When user gets locations configurations with "<request>"
    Then returns "<status>" status code
    And locations configurations are returned all items "<request>"

    Examples:
      | request                    | status          |
      | empty                      | InvalidArgument |
      | existing key and locations | OK              |
      | non existing key           | OK              |
