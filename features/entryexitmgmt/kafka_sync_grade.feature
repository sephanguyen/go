@quarantined
Feature: Sync data of grade table from mastermgmt to entryexitmgmt

    Scenario: Add record in grade table in mastermgmt and sync data to entryexitmgmt
        When a grade record is inserted in mastermgmt
        Then this grade record must be recorded in entryexitmgmt