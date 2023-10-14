@quarantined
Feature: Sync data of locations table from bob to entryexitmgmt

  Scenario: Add record in location in bob and sync data to entryexitmgmt
    When a location record is inserted in bob
    Then this location record must be recorded in entryexitmgmt