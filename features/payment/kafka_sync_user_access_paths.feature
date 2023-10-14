@wip @feature
Feature: User Access Paths sync data from Kafka

    Scenario: Add record in user access paths in bob and sync data to payment
        When a record is inserted in user access paths in bob
        Then the user access paths must be recorded in payment