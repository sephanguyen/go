Feature: Handle learning objectives created event

  @quarantined
  Scenario: Receive learning objectives created event
      Given an learning objectives created event
      When our system receives learning objectives created event
      Then our system must update study plan items correctly
