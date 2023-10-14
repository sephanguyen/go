# TODO refactor this
Feature: Run migration job add default usergroup for student & parent

  Scenario: Migrate run job to migrate assign user group to existed staff
    Given user group with role "<role>"
    And amount staff have "<usergroup>" old user group is "<school>"
    When admin choose the previous user group assign to "<amount>" of staff and run migration
    Then the specified staff must have the previous user group

    Examples:
      | usergroup               | school         | amount | role         |
      | USER_GROUP_TEACHER      | Manabie School | half   | Teacher      |
      | USER_GROUP_SCHOOL_ADMIN | Manabie School | none   | School Admin |
      | USER_GROUP_TEACHER      | Manabie School | none   | Teacher      |
      | USER_GROUP_SCHOOL_ADMIN | Manabie School | half   | School Admin |
