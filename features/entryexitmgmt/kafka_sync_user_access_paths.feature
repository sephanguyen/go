@quarantined
Feature: Sync data of user_access_paths table from bob to entryexitmgmt

  Scenario: Add record in user_access_paths in bob and sync data to entryexitmgmt
    When a user access paths record is inserted in bob
    Then this user access paths record must be recorded in entryexitmgmt