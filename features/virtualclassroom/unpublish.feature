Feature: Unpublish an Uploading Stream

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

  Scenario: Student unpublish and tries again but gets unpublished before status
    Given current lesson has "none" streaming learner
    And user signed as student who belong to lesson
    When user prepares to publish
    Then returns "OK" status code
    And user gets "none" publish status
    And current lesson "includes" streaming learner
    When user unpublish
    Then returns "OK" status code
    And user gets "none" unpublish status
    When user unpublish
    Then returns "OK" status code
    And user gets "unpublished before" unpublish status
    And current lesson "does not include" streaming learner