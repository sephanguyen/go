Feature: Health Check
    Scenario: Health Check OK
        Given everything is OK
        When health check endpoint called
        Then virtual classroom should return "OK" with status "SERVING" 