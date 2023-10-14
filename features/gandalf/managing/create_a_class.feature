@runsequence
Feature: Create a class

  Background:
    Given a signed in admin
    And a random number
    And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
    And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
    And admin inserts schools

  Scenario: teacher try to create a class of a school does not have config
    Given a signed in teacher
    And a CreateClassRequest
    And a "valid" schoolId in CreateClassRequest
    And a valid name in CreateClassRequest
    And default config for "class_plan" has "planName" is "School"
    And default config for "class_plan" has "planPeriod" is "2025-06-30 23:59:59"
    When user create a class
    Then Bob returns "OK" status code
    And Bob must create class from CreateClassRequest
    And class must has "plan_id" is "School"
    And class must has "plan_duration" is "0"
    And class must has "plan_expired_at" is "2025-06-30 23:59:59"
    And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And Bob must push msg "CreateClass" subject "Class.Upserted" to nats
    And Bob must push msg "ActiveConversation" subject "Class.Upserted" to nats

  Scenario: teacher try to create a class of a school does not have config
    Given a signed in teacher
    And a CreateClassRequest
    And a "valid" schoolId in CreateClassRequest
    And a valid name in CreateClassRequest
    And default config for "class_plan" has "planName" is "School"
    And default config for "class_plan" has "planPeriod" is "90"
    When user create a class
    Then Bob returns "OK" status code
    And Bob must create class from CreateClassRequest
    And class must has "plan_id" is "School"
    And class must has "plan_duration" is "90"
    And class must has "plan_expired_at" is "NULL"
    And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And Bob must push msg "CreateClass" subject "Class.Upserted" to nats
    And Bob must push msg "ActiveConversation" subject "Class.Upserted" to nats


  Scenario: teacher try to create a class of a school has config
    Given a signed in teacher
    And a CreateClassRequest
    And a "valid" schoolId in CreateClassRequest
    And a valid name in CreateClassRequest
    And this school has config "plan_id" is "School", "plan_expired_at" is "2025-06-30 23:59:59", "plan_duration" is 0
    When user create a class
    Then Bob returns "OK" status code
    And Bob must create class from CreateClassRequest
    And class must has "plan_id" is "School"
    And class must has "plan_duration" is "0"
    And class must has "plan_expired_at" is "2025-06-30 23:59:59"
    And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And Bob must push msg "CreateClass" subject "Class.Upserted" to nats
    And Bob must push msg "ActiveConversation" subject "Class.Upserted" to nats


  Scenario Outline: admin try to create a class with valid data
    Given a signed in "<role>" with school: 1
    And a CreateClassRequest
    And a "1" schoolId in CreateClassRequest
    And a valid name in CreateClassRequest
    And a 2 "teacher" ownerId with school id is 1 in CreateClassRequest
    And default config for "class_plan" has "planName" is "School"
    And default config for "class_plan" has "planPeriod" is "2025-06-30 23:59:59"
    When user create a class
    Then Bob returns "OK" status code
    And Bob must create class from CreateClassRequest
    And class must has "plan_id" is "School"
    And class must has "plan_duration" is "0"
    And class must has "plan_expired_at" is "2025-06-30 23:59:59"
    And class must have 2 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And Bob must push msg "CreateClass" subject "Class.Upserted" to nats
    And Bob must push msg "ActiveConversation" subject "Class.Upserted" to nats

    Examples:
      | role         |
      | admin        |
      | school admin |
