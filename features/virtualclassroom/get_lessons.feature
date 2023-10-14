Feature: Get Lessons API for Teacher Web

  Background:
    Given user signed in as school admin 
    When enter a school
    And have some centers
    And have some teacher accounts
    And have some student accounts
    And have some courses
    And have some student subscriptions
    And have some medias
    And an existing set of lessons
    And "enable" Unleash feature with feature name "Virtual_Classroom_SwitchNewDBConnection_Switch_DB_To_LessonManagement"

  Scenario Outline: teacher get list of lessons with paging
    Given "teacher" signin system
    When user gets list of lessons on page "<page>" with limit "<limit>"
    Then returns "OK" status code
    And returns a list of lessons with the correct page info

    Examples:
        | page | limit |
        | 1    | 5     |
        | 2    | 10    |
        | 3    | 10    |

  Scenario Outline: teacher get list of lessons using different time lookup and compare
    Given "teacher" signin system
    When user gets list of lessons using "<lesson_time_compare>" and "<time_lookup>"
    Then returns "OK" status code
    And returns a list of lessons based from the correct time

    Examples:
        | lesson_time_compare                  | time_lookup                                  |
        | LESSON_TIME_COMPARE_FUTURE           | TIME_LOOKUP_START_TIME                       |
        | LESSON_TIME_COMPARE_FUTURE_AND_EQUAL | TIME_LOOKUP_END_TIME                         |
        | LESSON_TIME_COMPARE_PAST             | TIME_LOOKUP_END_TIME_INCLUDE_WITHOUT_END_AT  |
        | LESSON_TIME_COMPARE_PAST_AND_EQUAL   | TIME_LOOKUP_END_TIME_INCLUDE_WITH_END_AT     |

  Scenario Outline: teacher get list of lessons using filters
    Given "teacher" signin system
    When user gets list of lessons using "<filter>" with "<filter_value>"
    Then returns "OK" status code
    And returns a list of correct lessons based from the filter

    Examples:
        | filter             | filter_value                       |
        | LOCATION_IDS       |                                    |
        | TEACHER_IDS        |                                    |
        | STUDENT_IDS        |                                    |
        | COURSE_IDS         |                                    |
        | LESSON_STATUS      | LESSON_SCHEDULING_STATUS_DRAFT     |
        | LESSON_STATUS      | LESSON_SCHEDULING_STATUS_PUBLISHED |
        | LESSON_STATUS      | LESSON_SCHEDULING_STATUS_COMPLETED |
        | LESSON_STATUS      | LESSON_SCHEDULING_STATUS_CANCELED  |
        | LIVE_LESSON_STATUS | LIVE_LESSON_STATUS_NOT_ENDED       |         
        | LIVE_LESSON_STATUS | LIVE_LESSON_STATUS_ENDED           |
        | FROM_DATE          |                                    |
        | TO_DATE            |                                    |
