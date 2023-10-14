@blocker
Feature: Kafka sync user basic info

    Scenario: Add record in user_basic_info in bob and sync data to invoicemgmt
        When a user basic info record is inserted in bob
        Then this user basic info record must be recorded in invoicemgmt