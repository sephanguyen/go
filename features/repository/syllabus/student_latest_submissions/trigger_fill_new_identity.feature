Feature: Trigger fill new identity

  Scenario: Trigger fill new identity
    Given <student_latest_submissions>valid study plan item in DB
    When insert a valid student latest submission
    Then student latest submission new identity filled