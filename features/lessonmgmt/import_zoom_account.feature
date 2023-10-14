Feature: Import Zoom Account
  Background:
    Given user signed in as school admin 
    And have some centers
  Scenario Outline: Import valid csv file
    Given user signed in as school admin
    When the zoom account request payload with "<row condition>"
    Then importing zoom account
    And returns "OK" status code

    Examples:
      | row condition          |
      | all valid rows         |
