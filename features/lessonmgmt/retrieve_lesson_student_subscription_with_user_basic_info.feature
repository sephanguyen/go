@quarantined
Feature: Admin Retrieve Student Subscriptions
    Background:
        Given some centers
    Scenario Outline: school admin retrieve student subscriptions 
        Given a signed in "school admin" with school: <school>
        And "enable" Unleash feature with feature name "Lesson_Student_UseUserBasicInfoTable"
        And 20 student subscriptions of school <school> are existed in DB with "2022-08-01T08:30:00+00:00" and "2022-08-10T10:30:00+00:00" using user basic info table
        When admin retrieve student subscriptions limit <limit>, offset <offset>, lesson date "2022-08-05T10:30:00+00:00" with filter "<keyword>", "<courserIDs>", "<grades>", "<classIDs>" and "<locationIDs>"
        Then returns "OK" status code
        And Bob must return list of <total> student subscriptions from "<rs_from_id>" to "<rs_to_id>" and page limit <limit> with next page offset "<next_offset>" and pre page offset "<pre_offset>" and around lesson date "2022-08-05T10:30:00+00:00" and filter "<has_filter>"

        Examples:
            | school | limit | offset | total | rs_from_id | rs_to_id | next_offset | pre_offset | has_filter | keyword      | courserIDs | grades | classIDs | locationIDs |
            | 111    | 5     | 0      | 20    | 0          | 4        | 4           | nil        |            |              |            |        |          |             |
            | 112    | 5     | 9      | 20    | 10         | 14       | 14          | 4          |            |              |            |        |          |             |
            | 113    | 5     | 18     | 20    | 19         | 19       | nil         | 13         |            |              |            |        |          |             |
            | 114    | 5     | 0      | 20    | 0          | 4        | 4           | nil        | yes        | student name |            |        |          |             |
            | 115    | 5     | 0      | 0     | 0          | nil      | nil         | nil        | yes        | nothing      |            |        |          |             |
            | 116    | 5     | 0      | 10    | 0          | 4        | 4           | nil        | yes        |              | 1          |        |          |             |
            | 117    | 5     | 0      | 20    | 0          | 4        | 4           | nil        | yes        |              | 1,2        |        |          |             |
            | 118    | 5     | 0      | 20    | 0          | 4        | 4           | nil        | yes        |              |            | 12     |          |             |
            | 119    | 5     | 0      | 0     | 0          | 4        | 4           | nil        | yes        |              |            | 10,11  |          |             |
            | 120    | 5     | 0      | 20    | 0          | 4        | 4           | nil        | yes        | student name | 1,2        | 12     |          |             |
            | 110    | 5     | 0      | 10    | 0          | 4        | 4           | nil        | yes        |              |            |        | 1,2      | 1,2         |
            | 127    | 5     | 0      | 0     | 0          | nil      | nil         | nil        | yes        |              |            |        | 1,2      | 3,4         |
            | 128    | 5     | 0      | 0     | 0          | nil      | nil         | nil        | yes        |              |            |        | 3,4      | 1,2         |
            | 129    | 5     | 0      | 0     | 0          | nil      | nil         | nil        | yes        |              |            |        | 3,4      | 3,4         |