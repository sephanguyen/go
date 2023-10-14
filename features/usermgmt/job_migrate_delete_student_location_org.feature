Feature: Run migration job to migrate delete student location org existed in our system

  @quarantined
  Scenario: Migrate delete student location org existed in our system
    Given students with location type org in our system
    When system run job to migrate delete student location org existed in our system
    Then existing students have default location are removed location type org