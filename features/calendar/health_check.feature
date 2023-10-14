Feature: Health Check
    Scenario: Health Check OK
        Given everything is OK
        When health check endpoint called
        Then calendar should return "OK" with status "SERVING" 