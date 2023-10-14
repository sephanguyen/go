Feature: Get ClassDo URL

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And has a ClassDo account
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario: teacher and student get the ClassDo link of a lesson
    Given "teacher" signin system
    And an existing lesson with a ClassDo link and owner
    When user gets the ClassDo link of a lesson
    Then returns "OK" status code
    And returns the expected ClassDo link

    When "student" signin system
    And user gets the ClassDo link of a lesson
    Then returns "OK" status code
    And returns the expected ClassDo link