@blocker
Feature: User sync data from Bob Kafka

  Scenario: Add record in user in bob and sync data to invoicemgmt
    When a user record is inserted in bob
    Then this user record must be recorded in invoicemgmt