@runsequence
Feature: Verify course after sync data

  Background:
    Given staff creates school admin account for partner manually
    And system has synced teacher from partner
    And "school admin" logins CMS
    And "teacher" logins Teacher App

  Scenario Outline: View course after sync data
    When system syncs course which belong to "<type>"
    Then school admin sees course on CMS
    And teacher "<result>" the course on Teacher App
    Examples:
      | type | result |
      | kid  | sees   |
    #| juku   | does not see | Because the ticket https://manabie.atlassian.net/browse/LT-4676

  Scenario: Edit course name by sync data
    Given system has synced course and class from partner
    And system syncs student account which associate with all available course-class
    When system syncs course with edited course name
    Then school admin sees edited course name on CMS
    And teacher sees edited course name on Teacher App
    And student sees edited course name on Learner App
