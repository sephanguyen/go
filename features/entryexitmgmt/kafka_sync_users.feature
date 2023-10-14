@quarantined
Feature: Sync data of users table from bob to entryexitmgmt

  Scenario: Add record in user in bob and sync data to entryexitmgmt
    When a user record is inserted in bob
    Then this user record must be recorded in entryexitmgmt