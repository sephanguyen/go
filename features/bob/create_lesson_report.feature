Feature: User Create LessonReports

  Background:
    Given a form's config for "individual lesson report" feature with school id
    And  a form's config for "group lesson report" feature with school id
    And  a form's config for "multi-version-feature-name" feature with school id
    And some teacher accounts with school id
    And some student accounts with school id
    And some live courses with school id
    And a live lesson

  Scenario: admin submit a new lesson report and after update it
    Given "staff granted role school admin" signin system
    When user submit a new lesson report
    Then returns "OK" status code
    And Bob have a new lesson report

    When user submit to update a lesson report
    Then returns "OK" status code
    And Bob have a lesson report

  Scenario: admin save a new draft lesson report and after update it
    Given "staff granted role school admin" signin system
    When user save a new draft lesson report
    Then returns "OK" status code
    And Bob have a new draft lesson report

    When user save to update a draft lesson report
    Then returns "OK" status code
    And Bob have a draft lesson report

  Scenario: admin save a new draft lesson report with multi version form config and after update it
    Given "staff granted role school admin" signin system
    When user save a new draft lesson report with multi version feature name is "multi-version-feature-name"
    Then returns "OK" status code
    And Bob have a new draft lesson report

    When user save to update a draft lesson report
    Then returns "OK" status code
    And Bob have a draft lesson report

  Scenario: admin submit a new lesson report and update scheduling status lesson to COMPLETED
    Given "staff granted role school admin" signin system
    And updates scheduling status in the lesson is "<scheduling_status_before>"
    And "enable" Unleash feature with feature name "BACKEND_Lesson_HandleUpdateStatusWhenUserSubmitLessonReport"
    When user submit a new lesson report
    Then returns "OK" status code
    And Bob have a new lesson report
    And "<must_have_event>" must have event from "<scheduling_status>" to "<scheduling_status_after>"
    And the lesson scheduling status updates to "<scheduling_status_after>"

    Examples:
      | scheduling_status_before           | scheduling_status_after            | scheduling_status                  | must_have_event |
      | LESSON_SCHEDULING_STATUS_DRAFT     | LESSON_SCHEDULING_STATUS_COMPLETED | LESSON_SCHEDULING_STATUS_PUBLISHED | yes             |
      | LESSON_SCHEDULING_STATUS_PUBLISHED | LESSON_SCHEDULING_STATUS_COMPLETED |                                    | yes             |
      | LESSON_SCHEDULING_STATUS_COMPLETED | LESSON_SCHEDULING_STATUS_COMPLETED |                                    | no              |
      | LESSON_SCHEDULING_STATUS_CANCELED  | LESSON_SCHEDULING_STATUS_CANCELED  |                                    | no              |

  Scenario: admin submit a new lesson report and after update it
    Given "staff granted role school admin" signin system
    When user submit a new lesson report
    Then returns "OK" status code
    And Bob have a new lesson report

    And lesson is locked by timesheet
    When user submit to update a lesson report
    Then returns "OK" status code
    And Bob have a lesson report with lesson is locked

  Scenario: admin submit a new lesson report with multi version form config and after update it
    Given "staff granted role school admin" signin system
    When user submit a new lesson report with multi version feature name is "multi-version-feature-name"
    Then returns "OK" status code
    And Bob have a new lesson report

    When user submit to update a lesson report
    Then returns "OK" status code

  Scenario: admin save a new draft lesson report and after update it
    Given "staff granted role school admin" signin system
    When user save a new draft lesson report
    Then returns "OK" status code
    And Bob have a new draft lesson report

    And lesson is locked by timesheet
    When user save to update a draft lesson report
    Then returns "OK" status code
    And Bob have a draft lesson report with lesson is locked
