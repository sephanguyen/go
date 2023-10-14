Feature: Import all-in-one csv file for test

  Scenario Outline: Import valid all-in-one csv file for test
    Given a valid request payload for importing all-in-one csv file for test
    When "school admin" import all-in-one csv file for test
    Then receives "OK" status code
    And the valid all-in-one csv file for test is imported successfully
