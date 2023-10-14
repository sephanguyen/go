@zeus
Feature: Message deduplication of Nats JetStream

  Scenario: Publisher publish multiple message with same Nats-Msg-Id
    Given A user with username is "Bob" and password is "m@n@bi3"
    When "Bob" publishes 10 message with subject "ActivityLog.Created" and same Nats-Msg-Id
    Then total activity log is inserted must be one