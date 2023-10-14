@quarantined
Feature: Handle assignments created event

    Background: valid course background
      And a valid course background

    Scenario: Receive assignments created event
      Given an assignments created event
      When our system receives assignments created event
      Then our system must update assignment study plan items correctly

    Scenario: Receive assignments created event when assign assignments to topic
      Given an assignments created event
      When our system receives assignments created event
      And assign assignments to topic
      Then our system must update assignment study plan items correctly