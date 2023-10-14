Feature: Publish event to signal graduation order successfully created

    Scenario: Publish order event log after submitting graduation order
        Given prepare valid order request for graduation
        And subscribe created order event log
        When "school admin" submit order
        Then an event must be published to signal graduation order submitted
