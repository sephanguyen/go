@wip @quarantine
Feature: Grade Message
    To receive message from Bob
    I need to subscribe to nats topic

    @wip @quarantined
    Scenario: bob import grade
        Given an grade valid request payload
        When "school admin" importing grade
        Then receives "OK" status code
        And payment save consistent record grade with bob
