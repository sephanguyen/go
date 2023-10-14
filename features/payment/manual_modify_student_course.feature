Feature: Manual modify student course

  Scenario Outline: Manual insert new student course success
    Given prepare data for manual insert student course
    When "school admin" submit manual modify student course request
    Then receives "OK" status code
