Feature: Upsert Study Plan Item

  Background:
    Given a signed in "school admin"
    And there are courses existed
    And there are study plan created in courses

  Scenario: Study plan item
    When user creates new "valid" study plan item
    Then returns "OK" status code
