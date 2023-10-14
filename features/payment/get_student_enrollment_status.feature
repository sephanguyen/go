Feature: Get Enrolled Status

  Scenario: Get organization enrolled status success
    Given prepare data for student enrollment status history
    When "school admin" get organization enrolled status
    Then check valid data get organization enrolled status
    And receives "OK" status code
   
  Scenario: Get student enrolled location success
    Given prepare data for student enrollment status history
    When "school admin" get student enrolled location
    Then check valid data get student enrolled location
    And receives "OK" status code

  Scenario: Get student enrollment status by location success
    Given prepare data for student enrollment status history
    When "school admin" get student enrollment status by location
    Then check valid data get student enrollment status by location
    And receives "OK" status code

