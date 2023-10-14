Feature: Health Check
    Scenario: Health Check OK
        Given everything is OK
        When health check endpoint called
        And returns "OK" status code