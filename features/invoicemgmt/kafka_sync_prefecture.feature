@blocker
Feature: Prefecture sync data from Bob Kafka

    Scenario: Add record prefecture from bob and sync data to invoicemgmt
        When a prefecture record is inserted in bob
        Then this prefecture record must be recorded in invoicemgmt
