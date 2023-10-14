Feature: Trigger fill new identity

  Scenario: Trigger fill new identity
    Given <student_event_logs>valid study plan item in DB
    When insert a valid student event log
    Then student event log new identity filled
