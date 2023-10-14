Feature: End Live Lesson

  Background:
    Given "school admin" signin system
    And a random number
    And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
    And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
    And admin inserts schools
    And some teacher accounts with school id
    And some student accounts with school id
    And some live courses with school id
    And some medias

    Given "teacher" signin system
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

    Given "student" signin system
    And his owned student UUID
    And a JoinClassRequest
    And a "valid" classCode in JoinClassRequest
    When user join a class
    Then returns "OK" status code
    And class must have 1 member is "USER_GROUP_TEACHER" and is owner "true" and status "CLASS_MEMBER_STATUS_ACTIVE"
    And class must have 1 member is "USER_GROUP_STUDENT" and is owner "false" and status "CLASS_MEMBER_STATUS_ACTIVE"
        #And student profile show "School" plan

    # And a list of valid topics
    And a list of courses are existed in DB of "above teacher"
    And a list of lessons are existed in DB of "above teacher" with start time "2030-06-05T00:00:00Z" and end time "2030-07-05T00:00:00Z"

    Given "current" user signed as teacher
    And user share a material with type is video in live lesson room
    Then returns "OK" status code
    And user get current material state of live lesson room is video

    When user "disables" chat of learners in a live lesson room
    Then returns "OK" status code
    And user gets learners chat permission to "disabled"

    When user request recording live lesson
    Then returns "OK" status code
    And user get current recording live lesson permission to start recording

    Given user signed as student who belong to lesson
    When user raise hand in live lesson room
    Then returns "OK" status code
    And user get hands up state

  Scenario: End live lesson
    Given teacher retrieve live lesson with start time "2030-07-01T00:00:00Z" and end time "2030-07-03T12:00:00Z"
    When teacher end one of the live lesson v1
    Then returns "OK" status code
    And bob must update lesson end at time v1
    And teacher retrieve live lesson with start time "2030-07-01T00:00:00Z" and end time "2030-07-03T12:00:00Z"
    And the ended lesson must have status completed v1
    And Bob must push msg "EndLiveLesson" subject "Lesson.Updated" to nats
    And user get current material state of live lesson room is empty
    And user get all learner's hands up states who all have value is off
    And live lesson is not recording
    And user gets learners chat permission to "enabled"
