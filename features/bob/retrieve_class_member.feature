@quarantined
Feature: Retrieve class member
    Background:
        Given "staff granted role school admin" signin system
        And a random number
        And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
        And admin inserts schools

        Given a signed in teacher
        And a CreateClassRequest
        And a "valid" schoolId in CreateClassRequest
        And a valid name in CreateClassRequest
        And this school has config "plan_id" is "School", "plan_expired_at" is "2025-06-30 23:59:59", "plan_duration" is 0
        When user create a class
        Then returns "OK" status code
        And Bob must create class from CreateClassRequest
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

    Scenario: unauthenticated user try to retrieve class member
        Given an invalid authentication token
        And a RetrieveClassMemberRequest
        When user retrieve class member
        Then returns "Unauthenticated" status code

    Scenario: user try to retrieve class member has 1 student
        Given a valid token of current teacher
        And a RetrieveClassMemberRequest
        And a valid classId in RetrieveClassMemberRequest
        When user retrieve class member
        Then returns "OK" status code
        And returns 1 student(s) and 1 teacher(s) RetrieveClassMemberResponse

    Scenario: user try to retrieve class member has 2 student
        Given a signed in student
        And a JoinClassRequest
        And a "valid" classCode in JoinClassRequest
        When user join a class
        Then returns "OK" status code
        And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And class must have 2 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And Bob must push msg "JoinClass" subject "Class.Upserted" to nats

        Given a valid token of current teacher
        And a RetrieveClassMemberRequest
        And a valid classId in RetrieveClassMemberRequest
        When user retrieve class member
        Then returns "OK" status code
        And returns 2 student(s) and 1 teacher(s) RetrieveClassMemberResponse
