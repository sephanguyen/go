Feature: Publish event to signal withdrawal order successfully created

    Scenario: Publish order event log after submitting withdrawal order
        Given prepare valid order request for withdrawal
        And subscribe created order event log
        When "school admin" submit order
        Then an event must be published to signal withdrawal order submitted

    Scenario: Publish order event log after voiding withdrawal order
        Given prepare valid order request for withdrawal with empty product
        And subscribe created order event log
        When "school admin" submit order
        And "school admin" void withdrawal order without products
        Then an event must be published to signal voiding of withdrawal order