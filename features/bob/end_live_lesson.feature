@quarantined
Feature: End Live Lesson

    Background:
        Given "staff granted role school admin" signin system
        And a random number
        And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
        And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
        And admin inserts schools

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
        And his owned student UUID
        And a JoinClassRequest
        And a "valid" classCode in JoinClassRequest
        When user join a class
        Then returns "OK" status code
        And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And class must have 1 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"
        
        And a list of courses are existed in DB of "above teacher"
        And a list of lessons are existed in DB of "above teacher" with start time "2030-06-05T00:00:00Z" and end time "2030-07-05T00:00:00Z"

    Scenario: End live lesson
        Given teacher retrieve live lesson with start time "2030-07-01T00:00:00Z" and end time "2030-07-03T12:00:00Z"
        When teacher end one of the live lesson
        Then returns "OK" status code
        And bob must update lesson end at time
        And teacher retrieve live lesson with start time "2030-07-01T00:00:00Z" and end time "2030-07-03T12:00:00Z"
        And the ended lesson must have status completed
        And Bob must push msg "EndLiveLesson" subject "Lesson.Updated" to nats

