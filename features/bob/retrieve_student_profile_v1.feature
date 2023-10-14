@quarantined @runsequence
Feature: Retrieve Student Profile

    Scenario: unauthenticated user retrieves student profile v1
        Given an invalid authentication token
            And an other student profile in DB
        When user retrieves student profile v1
        Then returns "Unauthenticated" status code

    # Scenario: user unregister retrieves student profile
    #     Given a valid authentication token
    #     And an other student profile in DB
    #     When user retrieves student profile v1
    #     Then returns "OK" status code
    #     And returns requested student profile v1

    Scenario: student retrieves some one else profile
        Given a signed in student
            And an other student profile in DB
        When user retrieves student profile v1
        Then returns "OK" status code
        And returns requested student profile v1

    Scenario: teacher get student profile in class
        Given "staff granted role school admin" signin system
        And a random number
        And a school name "S1", country "COUNTRY_VN", city "H·ªì Ch√≠ Minh", district "2"
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

        Given a signed in student
        And a JoinClassRequest
        And a "valid" classCode in JoinClassRequest
        When user join a class
        Then returns "OK" status code
        And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And class must have 1 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"

        Given a signed in student
        And a JoinClassRequest
        And a "valid" classCode in JoinClassRequest
        When user join a class
        Then returns "OK" status code
        And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And class must have 2 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"

        Given a valid token of current teacher

        When user retrieves student profile of classMembers v1
        Then returns "OK" status code
        And returns requested student profile v1


    Scenario: student get student profile in class
        Given "staff granted role school admin" signin system
        And a random number
        And a school name "S1", country "COUNTRY_VN", city "H·ªì Ch√≠ Minh", district "2"
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

        Given a signed in student
        And a JoinClassRequest
        And a "valid" classCode in JoinClassRequest
        When user join a class
        Then returns "OK" status code
        And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And class must have 1 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"

        Given a signed in student
        And a JoinClassRequest
        And a "valid" classCode in JoinClassRequest
        When user join a class
        Then returns "OK" status code
        And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
        And class must have 2 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"

        When user retrieves student profile of classMembers v1
        Then returns "OK" status code
        And returns requested student profile v1

    Scenario: teacher retrieves student profile
        Given a signed in teacher
        When teacher retrieves a "<kind of student>" student profile v1
        Then returns "OK" status code
        And returns requested student profile v1

    Examples:
      | kind of student      |
      | newly created        |
      | has signed in before |