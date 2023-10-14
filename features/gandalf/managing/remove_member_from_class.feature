Feature: Remove member from class

  Background:
    Given a signed in admin
    And a random number
    # a random number will be converted to string and will be appended to school_name
    # we need to do that because we have a unique constraint (school_name, country, city_id, district_id) in DB
    And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
    And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
    And these schools were inserted by admin

    And a class with school name "S1" and expired at "2150-12-12"
    And a class with school name "S2" and expired at "2150-12-12"

    #And some package plan available
    And 1 "student" join class with school name "S2"
    And 1 "teacher" join class with school name "S2"


  Scenario Outline: owner teacher try to remove a member
    Given a valid token of current teacher
    And a RemoveMemberRequest with class in school name "S2"
    And a "<user id>" in RemoveMemberRequest
    When user remove member from class
    Then Bob returns "OK" status code
    And class must have <number of active teacher> member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And class must have <number of inactive teacher> member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_INACTIVE"
    And class must have <number of active student> member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And class must have <number of inactive student> member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_INACTIVE"
    And Bob must push msg "LeaveClass-is_kicked=true" subject "Class.Upserted" to nats
    #And student profile show "<plan>" plan
    And Bob must store activity logs "remove" class member
    And Eureka must remove class members


    Examples:
      | user id         | number of active teacher | number of inactive teacher | number of active student | number of inactive student | plan   |
      | valid teacherId | 1                        | 1                          | 1                        | 0                          | School |
      | valid userId    | 2                        | 0                          | 0                        | 1                          | Trial  |

  Scenario Outline: admin try to remove member class with valid data
    Given a signed in "<role>" with school name "S2"
    And a RemoveMemberRequest with class in school name "S2"
    And a "<user id>" in RemoveMemberRequest
    When user remove member from class
    Then Bob returns "OK" status code
    And class must have <number of active teacher> member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And class must have <number of inactive teacher> member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_INACTIVE"
    And class must have <number of active student> member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And class must have <number of inactive student> member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_INACTIVE"
    And Bob must push msg "LeaveClass-is_kicked=true" subject "Class.Upserted" to nats
    #And student profile show "<plan>" plan
    And Bob must store activity logs "remove" class member
    And Eureka must remove class members

    Examples:
      | role         | user id         | number of active teacher | number of inactive teacher | number of active student | number of inactive student | plan   |
      | admin        | valid teacherId | 1                        | 1                          | 1                        | 0                          | School |
      | school admin | valid teacherId | 1                        | 1                          | 1                        | 0                          | School |
      | admin        | valid userId    | 2                        | 0                          | 0                        | 1                          | Trial  |
      | school admin | valid userId    | 2                        | 0                          | 0                        | 1                          | Trial  |

  Scenario Outline: admin try to remove 2 members
    Given a signed in "<role>" with school name "S2"
    And a RemoveMemberRequest with class in school name "S2"
    And a "<user id 1>" in RemoveMemberRequest
    And a "<user id 2>" in RemoveMemberRequest
    When user remove member from class
    Then Bob returns "OK" status code
    And class must have <number of active teacher> member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And class must have <number of inactive teacher> member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_INACTIVE"
    And class must have <number of active student> member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And class must have <number of inactive student> member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_INACTIVE"
    And Bob must push msg "LeaveClass-is_kicked=true" subject "Class.Upserted" to nats
    #And student profile show "<plan>" plan
    And Bob must store activity logs "remove" class member
    And Eureka must remove class members

    Examples:
      | role  | user id 1    | user id 2       | number of active teacher | number of inactive teacher | number of active student | number of inactive student | plan  |
      | admin | valid userId | valid teacherId | 1                        | 1                          | 0                        | 1                          | Trial |
