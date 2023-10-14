Feature: Retrieve classrooms by locations
  Background:
    When enter a school
    Given have some centers
    And have some classrooms

  Scenario Outline: Retrieve classrooms by locations
    Given user signed in as school admin
    When user gets classrooms of locations "7-4,7-20,7-21,3-20,3-21"
    Then returns "OK" status code
    And the list classrooms of these locations
