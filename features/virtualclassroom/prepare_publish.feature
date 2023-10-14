Feature: Prepare Publish an Uploading Stream

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing a virtual classroom session
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario: Student prepares to publish and tries again but gets prepared before status
    Given current lesson has "none" streaming learner
    And user signed as student who belong to lesson
    When user prepares to publish
    Then returns "OK" status code
    And user gets "none" publish status
    And current lesson "includes" streaming learner
    When user prepares to publish
    Then returns "OK" status code
    And user gets "prepared before" publish status
    And current lesson "includes" streaming learner

  Scenario: Student prepares to publish but has reached max limit
    Given current lesson has "max" streaming learner
    And user signed as student who belong to lesson
    When user prepares to publish
    Then returns "OK" status code
    And user gets "max limit" publish status
    And current lesson "does not include" streaming learner