@quarantined
Feature: Assign study plan to student

   Scenario: Assign study plan to student
      Given a valid course and study plan background
      When user assign study plan to a student
      Then returns "OK" status code
      And eureka must assign study plan to student