Feature: Admin Retrieve Assigned Student List
    Background:
        Given user signed in as school admin 
        And some centers
        And have some grades
        And a random number
        And a list of students from "id_0" to "id_5" purchase package 60 days with method "PURCHASE_METHOD_SLOT" of location "1" are existed in DB
        And a list of students from "id_5" to "id_10" purchase package 60 days with method "PURCHASE_METHOD_SLOT" of location "2" are existed in DB
        And a list of students from "id_10" to "id_20" purchase package 14 days with method "PURCHASE_METHOD_SLOT" of location "3" are existed in DB
        And a list of students from "id_20" to "id_25" purchase package 45 days with method "PURCHASE_METHOD_RECURRING" of location "1" are existed in DB
        And a list of students from "id_25" to "id_30" purchase package 45 days with method "PURCHASE_METHOD_RECURRING" of location "2" are existed in DB
        And "disable" Unleash feature with feature name "Lesson_Student_UseUserBasicInfoTable"

    Scenario Outline: school admin get assigned student list without filter
        Given "<role>" signin system
        When admin get assigned student list from "<purchase_method>" tab on page "<page_number>" with page limit of "<limit>"
        Then returns "OK" status code
        And must return assigned student list with correct total and offset values

        Examples:
            | role         | purchase_method           | page_number | limit |
            | school admin | PURCHASE_METHOD_SLOT      | 3           | 5     |
            | school admin | PURCHASE_METHOD_SLOT      | 4           | 5     |
            | school admin | PURCHASE_METHOD_SLOT      | 1           | 10    |
            | school admin | PURCHASE_METHOD_SLOT      | 2           | 10    |
            | school admin | PURCHASE_METHOD_RECURRING | 1           | 10    |
            | school admin | PURCHASE_METHOD_RECURRING | 3           | 5     |
            | teacher      | PURCHASE_METHOD_RECURRING | 2           | 10    |
            | teacher      | PURCHASE_METHOD_RECURRING | 1           | 20    |

    Scenario Outline: school admin get assigned student list with filter
        Given user signed in as school admin
        When admin get assigned student list from "<purchase_method>" tab with page limit of "<limit>" and filters "<key_word>", "<students>", "<courses>", "<centers>", "<locations>", "<status>", "<days_gap_start_date>", "<days_gap_end_date>"
        Then returns "OK" status code
        And must return correct assigned student list with "<key_word>", "<students>", "<courses>", "<centers>", "<locations>", "<status>"

        Examples:
            | purchase_method           | limit | key_word      | students                | days_gap_start_date | days_gap_end_date  | courses                     | centers | locations | status                        |
            | PURCHASE_METHOD_SLOT      | 10    | student name  | asg-std-8, asg-std-5    | -6                  | nil                | asg-course-8, asg-course-5  | 1,2     | 1         |                               |
            | PURCHASE_METHOD_RECURRING | 5     |               |                         | 3                   | nil                |                             | 2       |           | STUDENT_STATUS_OVER_ASSIGNED  |
            | PURCHASE_METHOD_SLOT      | 10    |               | asg-std-6               | 1                   | 60                 | asg-course-6                |         | 3         | STUDENT_STATUS_JUST_ASSIGNED  |
            | PURCHASE_METHOD_RECURRING | 10    | student name  | asg-std-2               | nil                 | nil                |                             | 2       |           | STUDENT_STATUS_OVER_ASSIGNED  |
                
