Feature: Health Check
    Scenario: Healtch Check OK
        Given everything is OK
        When health check endpoint called
        Then bob should return "OK" with status "SERVING"