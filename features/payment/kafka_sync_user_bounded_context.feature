Feature: Kafka sync user bounded context

  Scenario: Kafka sync user bounded context
    When admin inserts a user record to bob
    Then payment user table will be updated
