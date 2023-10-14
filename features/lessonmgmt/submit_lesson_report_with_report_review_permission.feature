Feature: User Submit LessonReports

  Background:
    When enter a school
    Given a form's config for "individual lesson report" feature with school id
    And a form's config for "group lesson report" feature with school id
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And "enable" Unleash feature with feature name "Lesson_LessonManagement_BackOffice_ReportReviewPermission"
    And a class with id prefix "prefix-class-id" and a course with id prefix "prefix-course-id"
  
  Scenario: Admin saves a new draft individual lesson report and update it
    Given user signed in as school admin
    And user has been granted "lesson.report.review" permission
    When user creates a new lesson with "individual" teaching method and all required fields
    Then returns "OK" status code
    And the lesson was created
    When user saves a new draft individual lesson report
    Then returns "OK" status code
    And the new saved draft lesson report existed in DB
  
  Scenario: Admin submits a new group lesson report
    Given user signed in as school admin
    And user has been granted "lesson.report.review" permission
    And user creates a new lesson with "group" teaching method and all required fields
    And returns "OK" status code
    And the lesson was created
    When user submits the created group lesson report
    Then returns "OK" status code
    And the new submitted lesson report existed in DB
