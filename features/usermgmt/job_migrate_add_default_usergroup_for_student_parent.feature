@quarantined
Feature: Run migration job add default usergroup for student & parent

  Scenario: Migrate add default usergroup for student & parent
    Given some students and parents without user group
    When system run job to migrate add default usergroup for student & parent
    Then previous students and parents have user group
