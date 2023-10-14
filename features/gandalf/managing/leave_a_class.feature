Feature: Leave a class
  Background:
    Given a signed in admin
    And a random number
      # a random number will be converted to string and will be appended to school_name
      # we need to do that because we have a unique constraint (school_name, country, city_id, district_id) in DB
    And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
    And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
    And these schools were inserted by admin

    #And some package plan available

    And a signed in teacher
    And a CreateClassRequest
    And a "valid" schoolId in CreateClassRequest
    And a valid name in CreateClassRequest
    And current school has config "plan_id" is "School", "plan_expired_at" is "2025-06-30 23:59:59", "plan_duration" is 0
    And a class created by current teacher with config "plan_id" is "School", "plan_expired_at" is "2025-06-30 23:59:59", "plan_duration" is  "0"


  Scenario: student try to leave a class with valid class id
    Given a signed in student
    And a JoinClassRequest
    And a "valid" classCode in JoinClassRequest
    And current student join class by using this class code
    And a LeaveClassRequest
    And a valid classId in LeaveClassRequest
    When user leave a class
    Then Bob returns "OK" status code
    And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And class must have 1 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_INACTIVE"
    And Bob must push msg "LeaveClass-is_kicked=false" subject "Class.Upserted" to nats
    #And student profile show "Trial" plan
    And Eureka must remove class members

  Scenario: teacher try to leave a class with valid class id
    Given a user signed in "teacher" with school name "S2"
    And a JoinClassRequest
    And a "valid" classCode in JoinClassRequest
    And current teacher join class by using this class code
    And a LeaveClassRequest
    And a valid classId in LeaveClassRequest
    When user leave a class
    Then Bob returns "OK" status code
    And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_INACTIVE"
    And Bob must push msg "LeaveClass-is_kicked=false" subject "Class.Upserted" to nats

