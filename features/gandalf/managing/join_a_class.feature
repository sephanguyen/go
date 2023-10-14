Feature: Join a class
  Background:
    Given a signed in admin
    And a random number
    And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
    And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
    And these schools were inserted by admin
    #And some package plan available
    And a class with school name "S2" and expired at "2025-06-30 23:59:59"

  Scenario: student try to join a class with valid class code
    Given a signed in student
    And a JoinClassRequest
    And a "valid" classCode in JoinClassRequest
    When user join a class
    Then Bob returns "OK" status code
    And JoinClassResponse must return "ClassId"
    And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And class must have 1 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And Bob must push msg "JoinClass" subject "Class.Upserted" to nats
    #And student profile show "School" plan
    #And student subscription must has "end_time" is "2025-06-30 23:59:59" with plan id is "School"
    And Eureka must add new class member
    # User try to join class again
    When user join a class
    Then Bob returns "AlreadyExists" status code

  Scenario Outline: users try to join with class with valid data and continue to join with class code already activated
    Given a class with school name "S2" and expired at "2150-12-12"
    #Scenario: users try to join a class with valid data
    And a user signed in "<role>" with school name "S2"
    And a JoinClassRequest
    And a "<class code>" classCode in JoinClassRequest
    When user join a class
    Then Bob returns "OK" status code
    And class must have <number> member is "<user group>" and is owner "<is owner>" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And Bob must push msg "JoinClass" subject "Class.Upserted" to nats
    And Eureka must add new class member
    #Scenario: continue to join with class code already activated
    When user join a class
    Then Bob returns "AlreadyExists" status code

    Examples:
      | role    | class code | number | user group         | is owner |
      | teacher | valid      | 2      | USER_GROUP_TEACHER | true     |
      | student | valid      | 1      | USER_GROUP_STUDENT | false    |

  Scenario Outline: student try to join a class with config
    Given a signed in teacher
    And a CreateClassRequest
    And a "valid" schoolId in CreateClassRequest
    And a valid name in CreateClassRequest
    And this school has config "plan_id" is "School", "plan_expired_at" is "<expired at>", "plan_duration" is <duration>
    And a class created by current teacher with config "plan_id" is "School", "plan_expired_at" is "<plan expired at>", "plan_duration" is  "<plan duration>"
    And a signed in student
    And a JoinClassRequest
    And a "valid" classCode in JoinClassRequest
    When user join a class
    Then Bob returns "OK" status code
    And class must have 1 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And Bob must push msg "JoinClass" subject "Class.Upserted" to nats
    #And student subscription must has "<type>" is "<end time>" with plan id is "School"
    And Eureka must add new class member

    Examples:
      | expired at          | duration | plan duration | plan expired at     | type         | end time            |
      | 2050-10-10 22:10:10 | 0        | 0             | 2050-10-10 22:10:10 | end_time     | 2050-10-10 22:10:10 |
      | NULL                | 0        | 0             | NULL                | end_time     | 2050-10-10 22:10:10 |
      | NULL                | 28       | 28            | NULL                | end_duration | 28                  |
