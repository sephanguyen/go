@quarantined
Feature: Update Student Product Status

  Scenario Outline: Cron Update student course by start time
    Given prepare data for create order recurring package
    And "school admin" submit order
    And receives "OK" status code
    And prepare data for update order recurring package
    And "school admin" submit order
    And receives "OK" status code
    And update recurring package success
    And prepare data for cron update student course
    When cron update student course run
    Then receives "OK" status code
