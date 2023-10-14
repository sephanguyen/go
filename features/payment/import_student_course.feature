Feature: Import Student Course

  Scenario Outline: Import student course valid csv file for create
    Given a student course valid request payload
    When "school admin" importing student course
    Then receives "OK" status code