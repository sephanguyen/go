@blocker
Feature: Audit configuration value change

  Scenario Outline: Audit configuration value change
    Given a "<type>" config key was inserted in DB
    When update the value of any "<type>" configuration
    Then the change be captured in the audit table of "<type>" configuration
    Examples:
      | type     |
      | internal |
      | external |
