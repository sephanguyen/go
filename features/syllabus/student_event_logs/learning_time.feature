Feature: Learning time
  Background:
    Given <learning_time>a signed in "student"
    
  Scenario: calculate learning time
    When student insert event log for learning time
    Then learning time is calculated
      And student's event log is stored
      And max score must be stored
