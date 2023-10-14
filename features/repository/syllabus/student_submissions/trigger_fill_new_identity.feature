Feature: Trigger fill new identity

  Scenario: Trigger fill new identity
    Given <student_submissions>valid study plan item in DB
    When insert a valid student submission
    Then student submission new identity filled