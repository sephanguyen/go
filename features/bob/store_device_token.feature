@quarantined
Feature: Store device token
    In order to use app
    As a user
    I need to store my device token

    Scenario: unauthenticated user try to store device token
        Given an invalid authentication token
        And a device token with device_token empty
        When user try to store device token
        Then returns "Unauthenticated" status code

    Scenario: admin try to store empty device token
        Given "staff granted role school admin" signin system
        And a device token with device_token empty
        When user try to store device token
        Then returns "InvalidArgument" status code
    
    Scenario: admin try to store device token
        Given "staff granted role school admin" signin system
        And a valid device token
        When user try to store device token
        # Then Bob must store the user's device token
        Then Bob must store the user's device token to user_device_tokens table
        And Bob must publish event to user_device_token channel

    Scenario: student try to store device token
        Given a signed in student
        And a valid device token
        When user try to store device token
        # Then Bob must store the user's device token
        Then Bob must store the user's device token to user_device_tokens table
        And Bob must publish event to user_device_token channel

    Scenario: student try to store empty device token
        Given a signed in student
        And a device token with device_token empty
        When user try to store device token
        Then returns "InvalidArgument" status code
