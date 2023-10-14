Feature: Run migration job to migrate set current school by grade in our system

#  @quarantined
  Scenario: Migrate set current school by grade in our system
    Given generate grade master
    And student info with school histories request and valid "one row current school"
    And "school admin" create new student account
    And generate school history without current school
    When system run job to migrate set current school by grade in our system
    Then existing school history with current school value set by grade value