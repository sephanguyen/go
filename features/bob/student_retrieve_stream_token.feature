@quarantined
Feature: student retrieve stream token
    In order for student to see stream
    As a student
    I need to retrieve a stream token

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
        And a JoinClassRequest
        And a "valid" classCode in JoinClassRequest
        When user join a class
        Then returns "OK" status code
        And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And class must have 1 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"
        #And student profile show "School" plan

    Scenario: student retrieve stream token
        Given a list of courses are existed in DB of "above teacher"
            And a student with valid lesson
        When student retrieve stream token
        Then returns "OK" status code
