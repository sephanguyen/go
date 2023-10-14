@wip @quarantine
Feature: Location sync data from Kafka

    @wip @quarantined
    Scenario: Add record location from bob and sync data to payment
        Given prepare location data
        When insert location data from bob
        Then payment must record location
