@quarantined
Feature: Handle ClassEvent
    Scenario: Handle JoinClass event
      Given an valid JoinClass event
      When send "ClassEvent" topic "Class.Upserted" to nats js
      Then our system must upsert ClassMember data correctly

    Scenario: Handle LeaveClass event
      Given an valid LeaveClass event
      When send "ClassEvent" topic "Class.Upserted" to nats js
      Then our system must update ClassMember data correctly