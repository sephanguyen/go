Feature: Insert Student Event Log

  Background:
    Given <student_event_logs>a signed in "student"
    
  Scenario: Student insert event log
    When student insert event log
    Then student event log must be created
    And student event log must be created with study_plan_item_id column
