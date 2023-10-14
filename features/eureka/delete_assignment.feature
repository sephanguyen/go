Feature: Upsert Assignment

   Scenario: Upsert assignments
      Given some assignments in db
      When user delete assignments
      Then returns "OK" status code
      And eureka must delete these assignments