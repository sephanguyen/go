Feature: create system notification
    Scenario: Consume Kafka messages and create System Notifications
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And some staffs with random roles and granted organization location of current organization
        And an "valid" upsert system notification kafka payload
        When publish event from kafka
        Then system notification data must be "created"
        When admin update content of system notification
        Then system notification data must be "updated"
        When admin update sent system notification to deleted
        Then system notification data must be "deleted"

    Scenario: Consume Kafka messages and do not create Invalid System Notifications
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And some staffs with random roles and granted organization location of current organization
        And an "invalid" upsert system notification kafka payload
        When publish event from kafka
        Then system notification data must be "not created"