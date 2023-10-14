@quarantined
Feature: Run migration job to create user group

  Scenario Outline: Migrate create user group
    Given some roles "<roles>" and locations to create user group
    When system run job to migrate create user group with userGroupName "<userGroupName>", roles "<roles>" and organization "<organization>"
    Then user group create successfully with userGroupName "<userGroupName>", roles "<roles>" and organization "<organization>"

    Examples:
      | roles        | userGroupName               | organization   |
      | School Admin | UserGroup School Admin test | MANABIE_SCHOOL |
      | HQ Staff     | User Group HQ Staff test    | MANABIE_SCHOOL |
      | Teacher      | User Group Teacher test     | MANABIE_SCHOOL |
