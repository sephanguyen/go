Feature: Import Student Class

  Scenario Outline: Import student class valid csv file for create
    Given a student class valid request payload
    When "school admin" importing student class for insert
    Then receives "OK" status code

  Scenario Outline: Import student class valid csv file for delete
    Given a student class valid request payload
    When "school admin" importing student class for delete
    Then receives "OK" status code