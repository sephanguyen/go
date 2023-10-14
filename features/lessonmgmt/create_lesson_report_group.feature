Feature: User Create LessonReports

  Background:
    When enter a school
    Given a form's config for "group lesson report" feature with school id
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias

  Scenario: admin save a new draft group lesson report and after update it
    Given user signed in as school admin
    And a class with id prefix "prefix-class-id" and a course with id prefix "prefix-course-id"
    When user creates a new lesson with "group" teaching method and all required fields
    Then returns "OK" status code
    And the lesson was created
    When user saves a new draft group lesson report
    Then returns "OK" status code
    And the new saved draft lesson report existed in DB
    When user saves to update a draft group lesson report
    Then returns "OK" status code
    And the new saved draft lesson report existed in DB

  Scenario: Admin submits a new group lesson report
    Given user signed in as school admin
    And a class with id prefix "prefix-class-id" and a course with id prefix "prefix-course-id"
    And user creates a new lesson with "group" teaching method and all required fields
    And returns "OK" status code
    And the lesson was created
    When user submits the created group lesson report
    Then returns "OK" status code
    And the new submitted lesson report existed in DB
 
  Scenario: Admin submits a new lesson report and update scheduling status lesson to COMPLETED
    Given user signed in as school admin
    And a class with id prefix "prefix-class-id" and a course with id prefix "prefix-course-id"
    And user creates a new lesson with "group" teaching method and all required fields
    And returns "OK" status code
    And the lesson was created
    And updates scheduling status in the lesson is "<scheduling_status_before>"
    And "enable" Unleash feature with feature name "BACKEND_Lesson_HandleUpdateStatusWhenUserSubmitLessonReport"
    When user submits the created group lesson report
    Then returns "OK" status code
    And the new submitted lesson report existed in DB
    And "<must_have_status>" must have event from "<scheduling_status>" to "<scheduling_status_after>"
    And the lesson scheduling status updates to "<scheduling_status_after>"

    Examples:
      | scheduling_status_before           | scheduling_status_after            | scheduling_status                  | must_have_status |
      | LESSON_SCHEDULING_STATUS_DRAFT     | LESSON_SCHEDULING_STATUS_COMPLETED | LESSON_SCHEDULING_STATUS_PUBLISHED | yes              |
      | LESSON_SCHEDULING_STATUS_PUBLISHED | LESSON_SCHEDULING_STATUS_COMPLETED |                                    | yes              |
      | LESSON_SCHEDULING_STATUS_COMPLETED | LESSON_SCHEDULING_STATUS_COMPLETED |                                    | no               |
      | LESSON_SCHEDULING_STATUS_CANCELED  | LESSON_SCHEDULING_STATUS_CANCELED  |                                    | no               |
  
    Scenario: validate attendance status when admin submits a new group report
      Given user signed in as school admin
      And a class with id prefix "prefix-class-id" and a course with id prefix "prefix-course-id"
      And user creates a new lesson with "group" teaching method and all required fields
      And returns "OK" status code
      And the lesson was created
      And updates scheduling status in the lesson is "<scheduling_status_before>"
      And "enable" Unleash feature with feature name "BACKEND_Lesson_HandleUpdateStatusWhenUserSubmitLessonReport"
      And "enable" Unleash feature with feature name "Lesson_LessonManagement_BackOffice_ValidationLessonBeforeCompleted"
      And students have attendance info "STUDENT_ATTEND_STATUS_EMPTY", "NOTICE_EMPTY", "REASON_EMPTY", ""
      When user submits the created group lesson report
      Then returns "OK" status code
      And the new submitted lesson report existed in DB
      And the lesson scheduling status updates to "<scheduling_status_after>"

      Examples:
        | scheduling_status_before           | scheduling_status_after            |
        | LESSON_SCHEDULING_STATUS_DRAFT     | LESSON_SCHEDULING_STATUS_PUBLISHED |
        | LESSON_SCHEDULING_STATUS_PUBLISHED | LESSON_SCHEDULING_STATUS_PUBLISHED |
        | LESSON_SCHEDULING_STATUS_COMPLETED | LESSON_SCHEDULING_STATUS_COMPLETED |
        | LESSON_SCHEDULING_STATUS_CANCELED  | LESSON_SCHEDULING_STATUS_CANCELED  |
