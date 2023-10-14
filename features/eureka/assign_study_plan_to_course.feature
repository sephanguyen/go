@quarantined
Feature: Assign study plan to course

   Scenario: Assign study plan to course
      Given a valid course and study plan background
      When user assign course with study plan 
      Then returns "OK" status code
      And eureka must assign study plan to course