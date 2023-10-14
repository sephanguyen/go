@blocker
Feature: Location sync data from Bob Kafka

    Scenario: Add record location from bob and sync data to invoicemgmt
        When a location record is inserted in bob
        Then this location record must be recorded in invoicemgmt
