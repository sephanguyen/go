@blocker
Feature: Student sync data from Bob Kafka

  Scenario: Add record in student in bob and sync data to invoicemgmt
    When a student record is inserted in bob
    Then this student record must be recorded in invoicemgmt