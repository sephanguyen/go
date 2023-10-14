Feature: Code Coverage Service for Git Branches
  @blocker
  Scenario: Set Coverage
    Given a git branch with no recorded coverage
    When client calls CreateTargetCoverage
    Then coverage amount is recorded in the database

  @blocker
  Scenario: Update Coverage
    Given a git branch with recorded coverage
    When client calls UpdateTargetCoverage
    Then coverage is updated in the database

  @blocker
  Scenario: Test Coverage
    Given created coverage
    When client calls TestCoverage with a "<amount>" coverage
    Then servers returns a "<statuscode>" result to the client
    Examples:
      | amount | statuscode       |
      | higher | FailPrecondition |
      | lower  | OK               |
