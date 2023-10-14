 Feature: Sync lesson data from bob to lessonmgmt database
  Background:
    Given user signed in as school admin 
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions

  Scenario: School admin can create a lesson without student
    Given user signed in as school admin
    When user creates a new lesson with all required fields in lessonmgmt
    Then returns "OK" status code
    And lesson data in lessonmgmt db has synced successfully
    And lesson teachers in lessonmgmt db has synced successfully

  Scenario: School admin can create a lesson with attendance info
    Given user signed in as school admin
    When user creates a new lesson with student attendance info "<attendance_status>", "<attendance_notice>", "<attendance_reason>", and "<attendance_note>"
    Then returns "OK" status code
    And lesson data in lessonmgmt db has synced successfully
    And lesson teachers in lessonmgmt db has synced successfully
    And lesson members in lessonmgmt db has synced successfully
    Examples:
      | attendance_status                 | attendance_notice | attendance_reason  | attendance_note  |
      | STUDENT_ATTEND_STATUS_ATTEND      | NOTICE_EMPTY      | REASON_EMPTY       |                  |
      | STUDENT_ATTEND_STATUS_ABSENT      | ON_THE_DAY        | PHYSICAL_CONDITION | medical exam     |
      | STUDENT_ATTEND_STATUS_LEAVE_EARLY | IN_ADVANCE        | REASON_OTHER       | personal errands |
      | STUDENT_ATTEND_STATUS_LATE        | NO_CONTACT        | REASON_OTHER       | traffic          |
