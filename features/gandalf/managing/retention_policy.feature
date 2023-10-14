@zeus
Feature: Nats JetStream retention policy

  Scenario: Message will be deleted after ack
    Given A user with username is "Bob" and password is "m@n@bi3"
    When "Bob" publishes 10 message with subject "ActivityLog.Created"
    And These activity log are created by Zeus
    Then Some message above must be deleted from stream "activitylog"