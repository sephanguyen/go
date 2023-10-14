Feature: Get Live Lessons by Location

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And "enable" Unleash feature with feature name "VirtualClassroom_Whitelist_CourseIDs_Get_Live_Lesson"
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario: Teacher gets live lessons
    Given existing a virtual classroom sessions
    When "teacher" signin system
    And user gets live lessons with filters
    Then returns "OK" status code
    And "teacher" receives live lessons that matches with the filters in the request

  Scenario: Teacher gets live lessons with only locations filter
    Given existing a virtual classroom sessions
    When "teacher" signin system
    And user gets live lessons
    Then returns "OK" status code
    And "teacher" receives live lessons that matches with the filters in the request

  Scenario: Student gets live lessons without whitelisted course IDs
    Given existing a virtual classroom sessions
    When user signed as student who belong to lesson
    And user gets live lessons with paging only
    Then returns "OK" status code
    And "student" receives live lessons