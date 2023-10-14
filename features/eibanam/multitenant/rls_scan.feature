Feature: Scan RLS

  Scenario: Check all tables have rls enabled
    When scanner scans on all tables
    Then those tables must has rls enabled

  Scenario: Check all tables have rls enabled and rls forced
    When scanner scans on all tables
    Then those tables must has rls enabled
    And those tables must has rls forced
