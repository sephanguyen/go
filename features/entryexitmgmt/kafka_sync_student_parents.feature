@quarantined
Feature: Sync data of student_parents table from bob to entryexitmgmt

  Scenario: Add record in student_parents in bob and sync data to entryexitmgmt
    When a student_parent record is inserted in bob
    Then this student_parent record must be recorded in entryexitmgmt