Feature: User updates (edits) lesson

  Background:
    Given user signed in as school admin
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And a form's config for "individual lesson report" feature with school id
    And have some classrooms

  Scenario Outline: School admin updates a lesson
    Given user signed in as school admin
    And an existing lesson in lessonmgmt
    When user updates "<field>" in the lesson lessonmgmt
    Then returns "OK" status code
    And the lesson was updated in lessonmgmt
    And Lessonmgmt must push msg "UpdateLesson" subject "Lesson.Updated" to nats

    Examples:
      | field             |
      | start time        |
      | end time          |
      | center id         |
      | teacher ids       |
      | student info list |
      | teaching medium   |
      | teaching method   |
      | material info     |
      | all fields        |

  Scenario Outline: Teacher updates a lesson
    Given user signed in as teacher
    And an existing lesson in lessonmgmt
    When user updates "<field>" in the lesson lessonmgmt
    Then returns "OK" status code
    And the lesson was updated in lessonmgmt
    And Lessonmgmt must push msg "UpdateLesson" subject "Lesson.Updated" to nats

    Examples:
      | field             |
      | start time        |
      | end time          |
      | center id         |
      | teacher ids       |
      | student info list |
      | teaching medium   |
      | teaching method   |
      | material info     |
      | all fields        |

  Scenario: School admin cannot update lessons with incorrect time
    Given user signed in as school admin
    And an existing lesson in lessonmgmt
    When user updates the lesson with start time later than end time in lessonmgmt
    Then returns "Internal" status code
#   And the lesson is not updated

  Scenario Outline: School admin cannot update lessons with empty required fields
    Given user signed in as school admin
    And an existing lesson in lessonmgmt
    When user updates the lesson with missing "<field>" in lessonmgmt
    Then returns "Internal" status code
#   And the lesson is not updated

    Examples:
      | field       |
      | center id   |
      | start time  |
      | end time    |
      | teacher ids |

  Scenario Outline: School admin can update a lesson with "group teaching method" with all required fields
    Given user signed in as school admin
    And a class with id prefix "<prefix-class-id>" and a course with id prefix "<prefix-course-id>"
    And user creates a new lesson with "group" teaching method and all required fields in lessonmgmt
    And returns "OK" status code
    And the lesson was created in lessonmgmt
    When user updates "<field>" in the lesson lessonmgmt
    Then returns "OK" status code
    And the lesson was updated in lessonmgmt
    Examples:
      | field             | prefix-class-id    | prefix-course-id    |
      | start time        | bdd-test-class-id- | bdd-test-course-id- |
      | end time          | bdd-test-class-id- | bdd-test-course-id- |
      | center id         | bdd-test-class-id- | bdd-test-course-id- |
      | teacher ids       | bdd-test-class-id- | bdd-test-course-id- |
      | student info list | bdd-test-class-id- | bdd-test-course-id- |
      | teaching medium   | bdd-test-class-id- | bdd-test-course-id- |
      | teaching method   | bdd-test-class-id- | bdd-test-course-id- |
      | material info     | bdd-test-class-id- | bdd-test-course-id- |
      | all fields        | bdd-test-class-id- | bdd-test-course-id- |

  Scenario Outline: School admin can edit lesson by saving draft from published lesson
    Given user signed in as school admin
    And user creates a "<lesson_status>" lesson with "<missing_fields>" in "<service>"
    And admin create a lesson report
    When user edit lesson by saving "<saving_type>" in "<service>"
    Then returns "OK" status code
    And the lesson was updated in lessonmgmt
    And lesson report state is "<state>"
    Examples:
      | lesson_status | missing_fields    | saving_type | state       | service    |
      | published     | none              | draft       | deleted     | lessonmgmt |
      | draft         | none              | published   | not deleted | lessonmgmt |
      | draft         | student info list | draft       | deleted     | lessonmgmt |

  Scenario Outline: School admin updates a lesson with date type
    Given user signed in as school admin
    And an existing lesson in lessonmgmt
    And a date "<date>", location "<location>", date type "<date type>", open time "<open time>", status "<status>", resource path "<resource path>" are existed in DB
    When user updated lesson location "<center>", start time "<start_time>", end time "<end_time>"
    Then returns "<code>" status code

    Examples:
      | code     | center | start_time                | end_time                  | date                      | location | date type | open time | status | resource path |
      | OK       | 1      | 2022-08-24T08:00:00+07:00 | 2022-08-24T16:00:00+07:00 | 2022-08-24T00:00:00+07:00 | 1        | regular   | 7:30      | draft  | -2147483648   |
      | Internal | 1      | 2022-08-18T08:00:00+07:00 | 2022-08-18T16:00:00+07:00 | 2022-08-18T00:00:00+07:00 | 1        | closed    | nil       | draft  | -2147483648   |
      | OK       | 2      | 2022-08-24T08:00:00+07:00 | 2022-08-24T16:00:00+07:00 |                           |          |           |           |        |               |
      | Internal | 3      | 2022-08-24T08:00:00+07:00 | 2022-08-24T16:00:00+07:00 | 2022-08-24T00:00:00+07:00 | 3        | closed    | nil       | draft  | -2147483648   |
      | Internal | 3      | 2022-08-24T04:00:00+07:00 | 2022-08-24T16:00:00+07:00 | 2022-08-24T00:00:00+07:00 | 3        | closed    | nil       | draft  | -2147483648   |
      | OK       | 3      | 2022-08-25T04:00:00+07:00 | 2022-08-25T16:00:00+07:00 | 2022-08-24T00:00:00+07:00 | 3        | closed    | nil       | draft  | -2147483648   |

  Scenario Outline: School admin updates a lesson with attendance info
    Given user signed in as school admin
    And an existing lesson with student attendance info "STUDENT_ATTEND_STATUS_EMPTY", "NOTICE_EMPTY", "REASON_EMPTY", and ""
    When user updates lesson student attendance info to "<attendance_status>", "<attendance_notice>", "<attendance_reason>", and "<attendance_note>"
    Then returns "OK" status code
    And the attendance info is updated
    Examples:
      | attendance_status                 | attendance_notice | attendance_reason  | attendance_note  |
      | STUDENT_ATTEND_STATUS_ATTEND      | NOTICE_EMPTY      | REASON_EMPTY       |                  |
      | STUDENT_ATTEND_STATUS_ABSENT      | ON_THE_DAY        | PHYSICAL_CONDITION | medical exam     |
      | STUDENT_ATTEND_STATUS_LEAVE_EARLY | IN_ADVANCE        | REASON_OTHER       | personal errands |
      | STUDENT_ATTEND_STATUS_LATE        | NO_CONTACT        | REASON_OTHER       | traffic          |

  Scenario Outline: School admin can create a lesson with classroom
    Given user signed in as school admin
    And an existing lesson with classroom
    When user updates lesson classroom with "<record_state>" record
    Then returns "<status>" status code
    And the lesson classrooms are "<update_state>"
    Examples:
      | record_state | status   | update_state |
      | existing     | OK       | updated      |
      | not existing | Internal | not updated  |

  Scenario Outline: School admin mark student as reallocation
    Given user signed in as school admin
    And an existing lesson with student attendance info "<previous_state>", "NOTICE_EMPTY", "REASON_EMPTY", and ""
    And returns "OK" status code
    And locks lesson
    When user marks student as reallocate
    Then returns "<status>" status code
    And student attendance status is "<update_state>"
    Examples:
      | status   | previous_state               | update_state                     |
      | OK       | STUDENT_ATTEND_STATUS_ABSENT | STUDENT_ATTEND_STATUS_REALLOCATE |
      | Internal | STUDENT_ATTEND_STATUS_ATTEND | STUDENT_ATTEND_STATUS_ATTEND     |
