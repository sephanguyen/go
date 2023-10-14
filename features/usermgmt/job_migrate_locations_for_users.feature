@quarantined
Feature: Run the job to migrate locations to empty users

  Scenario: Migrate locations to empty users
    Given some "<userType>" users are existed in Manabie system
    When we run migration specify "<school>" and pick "<location-type>" with "<userType>"
    Then existed "<userType>" with "<location-type>" must be assigned locations

     Examples:
      | school         | location-type     | userType |
      | Manabie School | empty location id | staff    |
      | Manabie School | one location id   | staff    |
      | Manabie School | empty location id | student  |
      | Manabie School | empty location id | parent   |

