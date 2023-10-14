Feature: Track appsmith events
    Scenario Outline: Save event to appsmith db
        Given a request with a "<auth header>" header
        When the track endpoint is called
        Then returns "<status code>" and "<response type>" response data
        Examples:
            | auth header | status code | response type |
            | wrong key   | 401         | error         |
            | wrong value | 401         | error         |
            | valid       | 200         | success       |
