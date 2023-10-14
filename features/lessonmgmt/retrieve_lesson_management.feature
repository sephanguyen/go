Feature: Admin Retrieve Lessons Management
    Background:
        Given user signed in as school admin 
        And some centers
        And "disable" Unleash feature with feature name "Lesson_Student_UseUserBasicInfoTable"

    Scenario Outline: school admin retrieve live lesson without filter
        Given a signed in "<role>" with school random id
        And a random number
        And a list of lessons management tab "<lesson_time>" from "<from_id>" to "<to_id>" of school are existed in DB
        When admin get live lesson management tab "<lesson_time>" with page "<limit>" and "<offset>"
        Then returns "OK" status code
        And Bob must return list of "<total>" lessons management from "<rs_from_id>" to "<rs_to_id>" and page limit "<limit>" with next page offset "<next_offset>" and pre page offset "<pre_offset>"

        Examples:
            | role         | from_id | to_id | limit | offset | total | rs_from_id | rs_to_id | next_offset | pre_offset | lesson_time        |
            | school admin | id_30   | id_40 | 5     | nil    | 10    | id_30      | id_34    | id_34       | nil        | LESSON_TIME_FUTURE |
            | school admin | id_30   | id_50 | 5     | id_40  | 20    | id_41      | id_45    | id_45       | id_35      | LESSON_TIME_FUTURE |
            | school admin | id_10   | id_20 | 5     | id_21  | 10    | nil        | nil      | nil         | nil        | LESSON_TIME_FUTURE |
            | school admin | id_30   | id_40 | 5     | nil    | 10    | id_30      | id_34    | id_34       | nil        | LESSON_TIME_PAST   |
            | school admin | id_30   | id_40 | 5     | id_34  | 10    | id_35      | id_39    | id_39       | nil        | LESSON_TIME_PAST   |
            | school admin | id_10   | id_20 | 5     | id_21  | 10    | nil        | nil      | nil         | nil        | LESSON_TIME_PAST   |
            | teacher      | id_30   | id_40 | 5     | nil    | 10    | id_30      | id_34    | id_34       | nil        | LESSON_TIME_FUTURE |
    
    Scenario Outline: school admin retrieve live lesson with filter
        Given a signed in "school admin" with school: 100
        And a list of lessons management of school 100 are existed in DB
        When admin get live lesson management tab "<lesson_time>" with filter "<key word>", "<date_range>", "<time_range>", "<teachers>", "<students>", "<courses>", "<center>", "<locations>", "<grade>", "<day_of_week>", "<scheduling_status>", "<classes>", "<grades_v2>"
        Then returns "OK" status code
        And Bob must return correct list lesson management tab "<lesson_time>" with "<key word>", "<date_range>", "<time_range>", "<teachers>", "<students>", "<courses>", "<center>", "<locations>", "<grade>", "<day_of_week>", "<scheduling_status>", "<classes>"

        Examples:
            | lesson_time        | key word     | date_range | time_range | day_of_week   | center   | teachers            | students            | courses   | grade | locations         | scheduling_status                                                     | classes         |grades_v2 |
            | LESSON_TIME_FUTURE | student name | yes        | yes        | 0,1,2,3,4,5,6 | center-1 | teacher-1,teacher-2 | student-1,student-2 | courser-1 | 12,13 | center-2,center-3 | LESSON_SCHEDULING_STATUS_PUBLISHED                                    | class-1,class-2 |Grade-1,Grade-3 |
            | LESSON_TIME_FUTURE | student name |            |            |               |          |                     |                     |           |       | center-1,center-2 | LESSON_SCHEDULING_STATUS_DRAFT                                        |                 |        |
            | LESSON_TIME_FUTURE |              | yes        |            |               |          |                     |                     |           |       | center-2,center-3 | LESSON_SCHEDULING_STATUS_COMPLETED                                    |                 |        |
            | LESSON_TIME_FUTURE |              | yes        |            |               |          |                     |                     |           | 0     |                   | LESSON_SCHEDULING_STATUS_PUBLISHED,LESSON_SCHEDULING_STATUS_COMPLETED |                 |        |
            | LESSON_TIME_FUTURE |              |            |            |               | center-1 |                     |                     |           |       | center-1          | LESSON_SCHEDULING_STATUS_DRAFT,LESSON_SCHEDULING_STATUS_CANCELED      |                 |        |
            | LESSON_TIME_FUTURE |              |            |            |               |          | teacher-1,teacher-2 |                     |           |       | center-1,center-3 | LESSON_SCHEDULING_STATUS_CANCELED                                     |                 |        |
            | LESSON_TIME_FUTURE |              |            |            |               |          |                     | student-1           |           |       |                   |                                                                       |                 |        |
            | LESSON_TIME_FUTURE |              |            |            |               |          |                     |                     | courser-1 |       |                   |                                                                       |                 |        |
            | LESSON_TIME_FUTURE |              |            |            |               |          |                     |                     |           | 12,13 | center-1          |                                                                       |                 |        |
            | LESSON_TIME_FUTURE | student name | yes        |            |               |          |                     |                     |           |       | center-2          |                                                                       |                 |        |
            | LESSON_TIME_FUTURE | student name | yes        | yes        | 0,1,2,3,4     |          |                     |                     |           |       | center-3          |                                                                       |                 |        |
            | LESSON_TIME_FUTURE | student name | yes        | yes        | 0,1,2,3,4     | center-1 | teacher-1           |                     |           |       | center-1,center-2 |                                                                       |                 |        |
            | LESSON_TIME_FUTURE | student name | yes        | yes        | 0,1,2,3,4     | center-1 | teacher-1,teacher-2 | student-1,student-2 |           |       |                   |                                                                       |                 |        |
            | LESSON_TIME_FUTURE | student name | yes        | yes        | 0,1,2,3,4     | center-1 | teacher-1,teacher-2 | student-1,student-2 | courser-1 |       | center-3          |                                                                       |                 |        |
            | LESSON_TIME_PAST   | student name | yes        | yes        | 0,1,2,3,4,5,6 | center-1 | teacher-1,teacher-2 | student-1,student-2 | courser-1 | 12,13 | center-2          |                                                                       |                 |        |
            | LESSON_TIME_PAST   | student name |            |            |               |          |                     |                     |           |       | center-1          | LESSON_SCHEDULING_STATUS_CANCELED                                     |                 |        |
            | LESSON_TIME_FUTURE | student name |            |            |               | center-1 |                     |                     |           |       |                   |                                                                       | class-1,class-2 |        |
