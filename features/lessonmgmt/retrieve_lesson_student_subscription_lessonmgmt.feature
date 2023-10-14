Feature: Admin Retrieve Student Subscriptions
    Background:
        Given user signed in as school admin
        And some centers
    Scenario Outline: school admin retrieve student subscriptions
        Given a signed in "school admin" with school: <school>
        And have some grades
        And "disable" Unleash feature with feature name "Lesson_Student_UseUserBasicInfoTable"
        And 20 student subscriptions of school <school> are existed in DB with "2022-08-01T08:30:00+00:00" and "2022-08-10T10:30:00+00:00"
        When admin retrieve student subscriptions limit <limit>, offset <offset>, lesson date "2022-08-05T10:30:00+00:00" with filter "<keyword>", "<courserIDs>", "<grades>", "<classIDs>", "<locationIDs>" and "<gradesV2>" in lessonmgmt
        Then returns "OK" status code
        And Lessonmgmt must return list of <total> student subscriptions from "<rs_from_id>" to "<rs_to_id>" and page limit <limit> with next page offset "<next_offset>" and pre page offset "<pre_offset>" and around lesson date "2022-08-05T10:30:00+00:00" and filter "<has_filter>"

        Examples:
            | school | limit | offset | total | rs_from_id | rs_to_id | next_offset | pre_offset | has_filter | keyword      | courserIDs | grades | classIDs | locationIDs | gradesV2 |
            | 111    | 5     | 0      | 20    | 0          | 4        | 4           | nil        |            |              |            |        |          |             | 1        |
            | 112    | 5     | 9      | 20    | 10         | 14       | 14          | 4          |            |              |            |        |          |             | 1      |
            | 113    | 5     | 18     | 20    | 19         | 19       | nil         | 13         |            |              |            |        |          |             | 1        |
            | 114    | 5     | 0      | 20    | 0          | 4        | 4           | nil        | yes        | student name |            |        |          |             |          |
            | 115    | 5     | 0      | 0     | 0          | nil      | nil         | nil        | yes        | nothing      |            |        |          |             |          |
            | 116    | 5     | 0      | 10    | 0          | 4        | 4           | nil        | yes        |              | 1          |        |          |             |          |
            | 117    | 5     | 0      | 20    | 0          | 4        | 4           | nil        | yes        |              | 1,2        |        |          |             |          |
            | 118    | 5     | 0      | 20    | 0          | 4        | 4           | nil        | yes        |              |            | 12     |          |             |          |
            | 119    | 5     | 0      | 0     | 0          | 4        | 4           | nil        | yes        |              |            | 10,11  |          |             |          |
            | 120    | 5     | 0      | 20    | 0          | 4        | 4           | nil        | yes        | student name | 1,2        | 12     |          |             |          |
            | 110    | 5     | 0      | 10    | 0          | 4        | 4           | nil        | yes        |              |            |        | 1,2      | 1,2         |          |
            | 127    | 5     | 0      | 0     | 0          | nil      | nil         | nil        | yes        |              |            |        | 1,2      | 3,4         |          |
            | 128    | 5     | 0      | 0     | 0          | nil      | nil         | nil        | yes        |              |            |        | 3,4      | 1,2         |          |
            | 129    | 5     | 0      | 0     | 0          | nil      | nil         | nil        | yes        |              |            |        | 3,4      | 3,4         |          |
    
    Scenario Outline: school admin retrieve student subscriptions with enrollment status
        Given a signed in "school admin" with school: <school>
        And have some grades
        And 20 student subscriptions of school <school> are existed in DB with enrollment status "<enrollment_status>" and range date from "2022-08-01T08:30:00+00:00" to "2022-08-10T10:30:00+00:00"
        And a list of lesson_student_subscription_access_path are existed in DB
        When admin retrieve student subscriptions limit <limit>, offset <offset>, lesson date "2022-08-05T10:30:00+00:00" with filter "<keyword>", "<courserIDs>", "<grades>", "<classIDs>", "<locationIDs>" and "<gradesV2>" in lessonmgmt
        Then returns "OK" status code
        And Lessonmgmt must return list of <total> student subscriptions from "<rs_from_id>" to "<rs_to_id>" and page limit <limit> with next page offset "<next_offset>" and pre page offset "<pre_offset>" and around lesson date "2022-08-05T10:30:00+00:00" and filter "<has_filter>"

        Examples:
            | school | enrollment_status                   | limit | offset | total | rs_from_id | rs_to_id | next_offset | pre_offset | has_filter | keyword | courserIDs | grades | classIDs | locationIDs |gradesV2 |
            | 121    | STUDENT_ENROLLMENT_STATUS_POTENTIAL | 5     | 0      | 20    | 0          | 4        | 4           | nil        |            |         |            |        |          |             |          |
            | 122    | STUDENT_ENROLLMENT_STATUS_ENROLLED  | 5     | 9      | 20    | 10         | 14       | 14          | 4          |            |         |            |        |          |             |          |
            | 124    | STUDENT_ENROLLMENT_STATUS_WITHDRAWN | 5     | 0      | 0     | nil        | nil      | nil         | nil        |            |         |            |        |          |             |          |
            | 125    | STUDENT_ENROLLMENT_STATUS_GRADUATED | 5     | 0      | 0     | nil        | nil      | nil         | nil        |            |         |            |        |          |             |          |
            | 126    | STUDENT_ENROLLMENT_STATUS_LOA       | 5     | 0      | 0     | nil        | nil      | nil         | nil        |            |         |            |        |          |             |          |
   
    Scenario Outline: school admin retrieve student subscriptions with lesson_date
        Given a signed in "school admin" with school: <school>
        And have some grades
        And 20 student subscriptions of school <school> are existed in DB with "<start_date>" and "<end_date>"
        And a list of lesson_student_subscription_access_path are existed in DB
        When admin retrieve student subscriptions limit <limit>, offset <offset>, lesson date "<lesson_date>" with filter "", "", "", "", "" and "" in lessonmgmt
        Then returns "OK" status code
        And Lessonmgmt must return list of <total> student subscriptions from "<rs_from_id>" to "<rs_to_id>" and page limit <limit> with next page offset "<next_offset>" and pre page offset "<pre_offset>" and around lesson date "<lesson_date>" and filter ""

        Examples:
            | school | limit | offset | total | rs_from_id | rs_to_id | next_offset | pre_offset | start_date                | end_date                  | lesson_date               |
            | 130    | 5     | 0      | 20    | 0          | 4        | 4           | nil        | 2022-08-01T08:30:00+00:00 | 2022-08-10T10:30:00+00:00 | 2022-08-05T10:30:00+00:00 |
            | 131    | 5     | 0      | 0     | nil        | nil      | nil         | nil        | 2022-08-01T08:30:00+00:00 | 2022-08-10T10:30:00+00:00 | 2022-07-31T10:00:00+00:00 |
            | 132    | 5     | 0      | 0     | nil        | nil      | nil         | nil        | 2022-08-01T08:30:00+00:00 | 2022-08-10T10:30:00+00:00 | 2022-08-11T00:00:00+00:00 |
