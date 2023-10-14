@runsequence
Feature: User creates lesson

  Background:
    When enter a school
    Given have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And have some classrooms

  Scenario: School admin can create a lesson with all required fields
    Given user signed in as school admin
    When user creates a new lesson with all required fields in lessonmgmt
    Then returns "OK" status code
    And the lesson was created in lessonmgmt
    And Lessonmgmt must push msg "CreateLessons" subject "Lesson.Created" to nats

  Scenario: School admin can create a lesson with "group teaching method" with all required fields
    Given user signed in as school admin
    And a class with id prefix "<prefix-class-id>" and a course with id prefix "<prefix-course-id>"
    When user creates a new lesson with "group" teaching method and all required fields in lessonmgmt
    Then returns "OK" status code
    And the lesson was created in lessonmgmt
      Examples:
      | prefix-class-id    | prefix-course-id    |
      | bdd-test-class-id- | bdd-test-course-id- |

  Scenario: School admin can save draft lesson by saving onetime
    Given user signed in as school admin
    When user creates a "<lesson_status>" lesson with "<missing_fields>" in "lessonmgmt"
    Then returns "OK" status code
    And the lesson was created in lessonmgmt
    Examples:
      | lesson_status  | missing_fields         |
      | draft          | none                   |
      | draft          | students               |
      | draft          | teachers               |
      | draft          | students and teachers  |
      | published      | students               |

    Scenario: School admin can save published lesson by saving onetime
    Given user signed in as school admin
    When user creates a "<lesson_status>" lesson with "<missing_fields>" in "lessonmgmt"
    Then returns "Internal" status code
    Examples:
      | lesson_status   | missing_fields         |
      | published       | teachers               |
      | published       | students and teachers  |

    Scenario: School admin can create a lesson and with day types
    Given user signed in as school admin
    And a date "<date>", location "<location>", date type "<date_type>", open time "<open_time>", status "<status>", resource path "<resource_path>" are existed in DB
    When user creates a new lesson with "<date>", "<location>", and other required fields in lessonmgmt
    Then returns "<status_code>" status code
    And the lesson was "<created_status>" in lessonmgmt
    Examples:
      | date                        | location  | date_type   | open_time  | status      | resource_path | status_code  | created_status |
      | 2022-08-09T07:30:00+07:00   | 1         |             |            |             |               | OK           | created        |
      | 2022-08-10T07:30:00+07:00   | 2         | regular     | 09:00      | draft       | -2147483648   | OK           | created        |
      | 2022-08-11T07:30:00+07:00   | 1         | seasonal    | 09:00      | draft       | -2147483648   | OK           | created        |
      | 2022-08-12T07:30:00+07:00   | 2         | closed      |            | draft       | -2147483648   | Internal     | not created    |

    Scenario: School admin can create a lesson with attendance info
    Given user signed in as school admin
    When user creates a new lesson with student attendance info "<attendance_status>", "<attendance_notice>", "<attendance_reason>", and "<attendance_note>"
    Then returns "OK" status code
    And the lesson was created in lessonmgmt
    And the attendance info is correct
    Examples:
      | attendance_status                 | attendance_notice | attendance_reason  | attendance_note  |
      | STUDENT_ATTEND_STATUS_ATTEND      | NOTICE_EMPTY      | REASON_EMPTY       |                  |
      | STUDENT_ATTEND_STATUS_ABSENT      | ON_THE_DAY        | PHYSICAL_CONDITION | medical exam     |
      | STUDENT_ATTEND_STATUS_LEAVE_EARLY | IN_ADVANCE        | REASON_OTHER       | personal errands |
      | STUDENT_ATTEND_STATUS_LATE        | NO_CONTACT        | REASON_OTHER       | traffic          |
      | STUDENT_ATTEND_STATUS_REALLOCATE  | NO_CONTACT        | REASON_OTHER       | traffic          |

  Scenario: School admin can create a lesson with classrooms
    Given user signed in as school admin
    When user creates a new lesson with "<record_state>" classrooms
    Then returns "<status>" status code
    And the lesson was "<create_status>" in lessonmgmt
    And the classrooms are "<record_state>" in the lesson
    Examples:
      | record_state  | status    | create_status |
      | existing      | OK        | created       |
      | not existing  | Internal  | not created   |
