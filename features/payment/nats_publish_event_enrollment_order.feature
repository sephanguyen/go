@quarantined
Feature: Publish event to signal enrollment order successfully created

    Scenario Outline: Publish order event log after creating enrollment order
        Given prepare valid order request for enrollment
        And subscribe created order event log
        When "school admin" submit order
        Then an event must be published and enrollment status is updated "<result>"

        Examples:
            | result         |
            | successfully   |