@quarantined
Feature: Leave a class
    Background:
        Given "staff granted role school admin" signin system
        And a random number
        And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
        And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
        And admin inserts schools

        #And some package plan available

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
        And Bob must push msg "CreateClass" subject "Class.Upserted" to nats


        Given a signed in student
        And a JoinClassRequest
        And a "valid" classCode in JoinClassRequest
        When user join a class
        Then returns "OK" status code
        And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And class must have 1 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And Bob must push msg "JoinClass" subject "Class.Upserted" to nats
        #And student profile show "School" plan

    Scenario: unauthenticated user try to leave a class
        Given an invalid authentication token
        And a LeaveClassRequest
        When user leave a class
        Then returns "Unauthenticated" status code

    Scenario: user try to leave a class with valid class id
        And a LeaveClassRequest
        And a valid classId in LeaveClassRequest
        When user leave a class
        Then returns "OK" status code
        And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And class must have 1 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_INACTIVE"
        And Bob must push msg "LeaveClass-is_kicked=false" subject "Class.Upserted" to nats
        #And student profile show "Trial" plan

    Scenario: user try to leave a class without class id
        Given a signed in student
        And a LeaveClassRequest
        When user leave a class
        Then returns "InvalidArgument" status code
