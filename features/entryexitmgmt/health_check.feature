@quarantined
Feature: Health Check
    Scenario: Health Check OK
        Given everything is OK
        When health check endpoint called
        Then entryexitmgmt should return "OK" with status "SERVING"