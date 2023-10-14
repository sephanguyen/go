@blocker
Feature: User Access Paths sync data from Bob Kafka

  Scenario: Add record in user access path in bob and sync data to invoicemgmt
    When a user access path record is inserted in bob
    Then this user access path record must be recorded in invoicemgmt