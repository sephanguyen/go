@quarantined
Feature: Retrieve Lessons
    Background:
        Given "staff granted role school admin" signin system
        And a random number
        And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
        And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
        And admin inserts schools
        # And a list of valid topics
        # And admin inserts a list of valid topics

        Given a signed in teacher
        And a CreateClassRequest
        And a "valid" schoolId in CreateClassRequest
        And a valid name in CreateClassRequest
        And this school has config "plan_id" is "School", "plan_expired_at" is "2025-06-30 23:59:59", "plan_duration" is 0
        When user create a class
        Then returns "OK" status code
        And Bob must create class from CreateClassRequest
        And class must has "plan_id" is "School"
        And class must has "plan_duration" is "0"
        And class must has "plan_expired_at" is "2025-06-30 23:59:59"
        And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"

        Given a signed in student
        And a JoinClassRequest
        And a "valid" classCode in JoinClassRequest
        When user join a class
        Then returns "OK" status code
        And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And class must have 1 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"
        #And student profile show "School" plan


    Scenario Outline: student retrieve live lesson with start time and end time
        # Given a list of valid topics
        And a list of courses are existed in DB of "above teacher"
        And a list of lessons are existed in DB of "above teacher" with start time "<lesson_start_time>" and end time "<lesson_end_time>"
        And current student assigned to above lessons
        When student retrieve live lesson with start time "<retrieve_start_time>" and end time "<retrieve_end_time>"
        Then returns "OK" status code
        And Bob must return "correct" live lesson for student
        Examples:
            | lesson_start_time    | lesson_end_time      | retrieve_start_time  | retrieve_end_time    |
            | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-06T00:00:00Z | 2020-07-08T00:00:00Z |
            | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-06T00:00:00Z |
            | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-08T00:00:00Z |
            | 2020-07-05T12:00:00Z | 2020-07-07T00:00:00Z | 2020-07-05T12:01:00Z | 2020-07-05T12:02:00Z |

    Scenario Outline: student retrieve live lesson with invalid start time and end time
        # Given a list of valid topics
        And a list of courses are existed in DB of "above teacher"
        And a list of lessons are existed in DB of "above teacher" with start time "<lesson_start_time>" and end time "<lesson_end_time>"
        When student retrieve live lesson with start time "<retrieve_start_time>" and end time "<retrieve_end_time>"
        Then returns "OK" status code
        And Bob must return "empty" live lesson for student
        Examples:
            | lesson_start_time    | lesson_end_time      | retrieve_start_time  | retrieve_end_time    |
            | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-01T00:00:00Z | 2020-07-04T00:00:00Z |
            | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-08T00:00:00Z | 2020-07-10T00:00:00Z |

    Scenario Outline: teacher retrieve live lesson with start time and end time
        # Given a list of valid topics
        And a list of courses are existed in DB of "above teacher"
        And a list of lessons are existed in DB of "above teacher" with start time "<lesson_start_time>" and end time "<lesson_end_time>"
        When teacher retrieve live lesson with start time "<retrieve_start_time>" and end time "<retrieve_end_time>"
        Then returns "OK" status code
        And Bob must return "correct" live lesson for teacher
        Examples:
            | lesson_start_time    | lesson_end_time      | retrieve_start_time  | retrieve_end_time    |
            | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-06T00:00:00Z | 2020-07-08T00:00:00Z |
            | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-06T00:00:00Z |
            | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-08T00:00:00Z |
            | 2020-07-05T12:00:00Z | 2020-07-07T00:00:00Z | 2020-07-05T12:01:00Z | 2020-07-05T12:02:00Z |

    Scenario Outline: teacher retrieve live lesson with invalid start time and end time
        # Given a list of valid topics
        And a list of courses are existed in DB of "above teacher"
        And a list of lessons are existed in DB of "above teacher" with start time "<lesson_start_time>" and end time "<lesson_end_time>"
        When teacher retrieve live lesson with start time "<retrieve_start_time>" and end time "<retrieve_end_time>"
        Then returns "OK" status code
        And Bob must return "empty" live lesson for teacher
        Examples:
            | lesson_start_time    | lesson_end_time      | retrieve_start_time  | retrieve_end_time    |
            | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-01T00:00:00Z | 2020-07-04T00:00:00Z |
            | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-08T00:00:00Z | 2020-07-10T00:00:00Z |

    Scenario Outline: teacher retrieve live lesson multy course with start time and end time
        # Given a list of valid topics
        And a list of courses are existed in DB of "above teacher"
        And a list of lessons are existed in DB of "<lesson_opt>" with start time "<lesson_start_time>" and end time "<lesson_end_time>"
        When teacher retrieve live lesson by courseID "<course_id>" with start time "<retrieve_start_time>" and end time "<retrieve_end_time>"
        Then returns "OK" status code
        And Bob must return "<result>" live lesson for teacher
        Examples:
            | lesson_opt                                | course_id             | lesson_start_time    | lesson_end_time      | retrieve_start_time  | retrieve_end_time    | result  |
            | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-06T00:00:00Z | 2020-07-08T00:00:00Z | empty   |
            | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-06T00:00:00Z | empty   |
            | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-08T00:00:00Z | empty   |
            | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T12:00:00Z | 2020-07-07T00:00:00Z | 2020-07-05T12:01:00Z | 2020-07-05T12:02:00Z | empty   |
            | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-06T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-06T00:00:00Z | correct |
            | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T12:00:00Z | 2020-07-07T00:00:00Z | 2020-07-05T12:01:00Z | 2020-07-05T12:02:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-06T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-06T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T12:00:00Z | 2020-07-07T00:00:00Z | 2020-07-05T12:01:00Z | 2020-07-05T12:02:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-06T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-06T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T12:00:00Z | 2020-07-07T00:00:00Z | 2020-07-05T12:01:00Z | 2020-07-05T12:02:00Z | correct |

    Scenario Outline: teacher retrieve live lesson multy course with invalid start time and end time
        # Given a list of valid topics
        And a list of courses are existed in DB of "above teacher"
        And a list of lessons are existed in DB of "<lesson_opt>" with start time "<lesson_start_time>" and end time "<lesson_end_time>"
        When teacher retrieve live lesson by courseID "<course_id>" with start time "<retrieve_start_time>" and end time "<retrieve_end_time>"
        Then returns "OK" status code
        And Bob must return "<result>" live lesson for teacher
        Examples:
    | lesson_opt                               | course_id             | lesson_start_time    | lesson_end_time      | retrieve_start_time  | retrieve_end_time    | result |
    | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-01T00:00:00Z | 2020-07-04T00:00:00Z | empty  |
    | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-08T00:00:00Z | 2020-07-10T00:00:00Z | empty  |
    | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-01T00:00:00Z | 2020-07-04T00:00:00Z | empty  |
    | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-08T00:00:00Z | 2020-07-10T00:00:00Z | empty  |
    | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-01T00:00:00Z | 2020-07-04T00:00:00Z | empty  |
    | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-08T00:00:00Z | 2020-07-10T00:00:00Z | empty  |
    | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-01T00:00:00Z | 2020-07-04T00:00:00Z | empty  |
    | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-08T00:00:00Z | 2020-07-10T00:00:00Z | empty  |

    Scenario Outline: student retrieve live lesson multy course with start time and end time
        # Given a list of valid topics
        And a list of courses are existed in DB of "above teacher"
        And a list of lessons are existed in DB of "<lesson_opt>" with start time "<lesson_start_time>" and end time "<lesson_end_time>"
        And current student assigned to above lessons
        When student retrieve live lesson by courseID "<course_id>" with start time "<retrieve_start_time>" and end time "<retrieve_end_time>"
        Then returns "OK" status code
        And Bob must return "<result>" live lesson for student
        Examples:
            | lesson_opt                                | course_id             | lesson_start_time    | lesson_end_time      | retrieve_start_time  | retrieve_end_time    | result  |
            | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-06T00:00:00Z | 2020-07-08T00:00:00Z | empty   |
            | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-06T00:00:00Z | empty   |
            | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-08T00:00:00Z | empty   |
            | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T12:00:00Z | 2020-07-07T00:00:00Z | 2020-07-05T12:01:00Z | 2020-07-05T12:02:00Z | empty   |
            | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-06T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-06T00:00:00Z | correct |
            | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T12:00:00Z | 2020-07-07T00:00:00Z | 2020-07-05T12:01:00Z | 2020-07-05T12:02:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-06T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-06T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T12:00:00Z | 2020-07-07T00:00:00Z | 2020-07-05T12:01:00Z | 2020-07-05T12:02:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-06T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-06T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-04T00:00:00Z | 2020-07-08T00:00:00Z | correct |
            | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T12:00:00Z | 2020-07-07T00:00:00Z | 2020-07-05T12:01:00Z | 2020-07-05T12:02:00Z | correct |

    Scenario Outline: student retrieve live lesson multy course with invalid start time and end time
        # Given a list of valid topics
        And a list of courses are existed in DB of "above teacher"
        And a list of lessons are existed in DB of "<lesson_opt>" with start time "<lesson_start_time>" and end time "<lesson_end_time>"
        And current student assigned to above lessons
        When student retrieve live lesson by courseID "<course_id>" with start time "<retrieve_start_time>" and end time "<retrieve_end_time>"
        Then returns "OK" status code
        And Bob must return "<result>" live lesson for student
        Examples:
            | lesson_opt                                | course_id             | lesson_start_time    | lesson_end_time      | retrieve_start_time  | retrieve_end_time    | result |
            | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-01T00:00:00Z | 2020-07-04T00:00:00Z | empty  |
            | above teacher and belong to single course | course-live-teacher-7 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-08T00:00:00Z | 2020-07-10T00:00:00Z | empty  |
            | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-01T00:00:00Z | 2020-07-04T00:00:00Z | empty  |
            | above teacher and belong to single course | course-live-teacher-4 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-08T00:00:00Z | 2020-07-10T00:00:00Z | empty  |
            | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-01T00:00:00Z | 2020-07-04T00:00:00Z | empty  |
            | above teacher and belong to multy course  | course-live-teacher-5 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-08T00:00:00Z | 2020-07-10T00:00:00Z | empty  |
            | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-01T00:00:00Z | 2020-07-04T00:00:00Z | empty  |
            | above teacher and belong to multy course  | course-live-teacher-6 | 2020-07-05T00:00:00Z | 2020-07-07T00:00:00Z | 2020-07-08T00:00:00Z | 2020-07-10T00:00:00Z | empty  |