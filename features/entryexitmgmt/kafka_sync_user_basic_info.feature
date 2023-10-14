@major
Feature: Kafka sync user basic info from bob to entryexitmgmt

    Scenario: Add record in user_basic_info in bob and sync data to entryexitmgmt
        When a user basic info record is inserted in bob
        Then this user basic info record must be recorded in entryexitmgmt