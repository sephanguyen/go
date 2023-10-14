@quarantined
Feature: Sync data of students table from bob to entryexitmgmt

  Scenario: Add record in student in bob and sync data to entryexitmgmt
    When a student record is inserted in bob
    Then this student record must be recorded in entryexitmgmt