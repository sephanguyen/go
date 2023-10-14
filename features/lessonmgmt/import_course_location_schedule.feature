Feature: Import Course Location Schedule
  Background:
    Given user signed in as school admin 
    And have some centers
  Scenario Outline: Import valid csv file
    Given user signed in as school admin
    When the course location schedule request payload with "<row condition>"
    Then importing course location schedule
    And returns "OK" status code

    Examples:
      | row condition          |
      | all valid rows         |
