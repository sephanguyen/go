Feature: Health Check
    Scenario: Health Check OK
        Given everything is OK
        When health check endpoint called
        Then lesson mgmt should return "OK" with status "SERVING"