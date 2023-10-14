Feature: Login on CMS

  @wip
  Scenario: Login as manabie super admin
    When super admin logins on CMS
    Then super admin sees all data of all organization on CMS

  @wip
  Scenario: Display correct data when login school admin account from different organization
    Given "school admin 1" logins on CMS
    And "school admin 1" only interacts with content from "organization 1" on CMS
    When "school admin 1" logs out on CMS
    And "school admin 2" logins on CMS
    Then "school admin 2" only interacts with content from "organization 2" on CMS

  @wip
  Scenario: Display correct data when login teacher account from different organization
    Given "teacher 1" logins on CMS
    And "teacher 1" only interacts with content from "organization 1" on CMS
    When "teacher 1" logs out on CMS
    And "teacher 2" logins on CMS
    Then "teacher 2" only interacts with content from "organization 2" on CMS