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

  Scenario: Admin cannot submits a new individual lesson report when missing permission report.review
    Given user signed in as school admin
    And a class with id prefix "prefix-class-id" and a course with id prefix "prefix-course-id"
    And user creates a new lesson with "individual" teaching method and all required fields
    And returns "OK" status code
    And the lesson was created
    When user submits the created individual lesson report
    Then returns "Internal" status code

  Scenario: Admin cannot submits a new group lesson report when missing permission report.review
    Given user signed in as school admin
    And user creates a new lesson with "group" teaching method and all required fields
    And returns "OK" status code
    And the lesson was created
    When user submits the created group lesson report
    Then returns "Internal" status code

