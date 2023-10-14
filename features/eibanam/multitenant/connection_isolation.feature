Feature: Isolate user connection

  Scenario: Check coccurent connection isolation
    Given a random table in db
    And 1000 record with different resource path
    When rls is enable for table
    Then user can only fetch their data