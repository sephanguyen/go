@quarantined @runsequence
Feature: Retrieve Courses

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

    Scenario: unauthenticated user try to retrieve courses
        Given an invalid authentication token
        And an invalid student UUID
        When student retrieves courses
        Then returns "Unauthenticated" status code

    Scenario: student try to retrieve courses
        Given a signed in student
        And a list of courses are existed in DB of "manabie"
        When student retrieves courses
        Then returns "OK" status code
        And returns a list of courses of "manabie"

    Scenario: student try to retrieve courses from teacher and manabie content
        Given a list of courses are existed in DB of "manabie"
        Given a list of courses are existed in DB of "above teacher"
        When student retrieves courses
        Then returns "OK" status code
        And returns a list of courses of "manabie & above teacher"

    Scenario: student try to retrieve assigned courses
        Given a list of courses are existed in DB of "manabie"
        Given a list of courses are existed in DB of "above teacher"
        When student retrieves assigned courses
        Then returns "OK" status code
        And returns a list of courses of "manabie & above teacher"

    Scenario: student try to retrieve assigned courses with retrieve course end point
        Given a list of courses are existed in DB of "manabie"
        Given a list of courses are existed in DB of "above teacher"
        When student retrieves assigned courses with retrieve course end point
        Then returns "OK" status code
        And returns a list of courses of "manabie & above teacher"

    Scenario: student try to retrieve courses in current class
        Given a list of courses are existed in DB of "manabie"
        Given a list of courses are existed in DB of "above teacher"
        When student retrieves courses in current class
        Then returns "OK" status code
        And returns a list of courses in current class

    Scenario: student try to retrieve assigned live courses
        Given a list of courses are existed in DB of "manabie"
        Given a list of courses are existed in DB of "above teacher"
        When student retrieves assigned live courses
        Then returns "OK" status code
        And returns a list of courses of "manabie & above teacher"

    Scenario: student try to retrieve removed live courses
        Given a list of courses are existed in DB of "manabie"
        Given a list of courses are existed in DB of "above teacher"
        Given student's class is removed from course
        When student retrieves assigned live courses belong to current class
        Then returns "OK" status code
        And returns empty list of course

    Scenario: student try to retrieve course with ids
        Given a list of courses are existed in DB of "manabie"
        When student retrieves courses with ids
        Then returns "OK" status code
        And returns a list of courses from requested ids

    Scenario: student try to retrieve course with existing lesson
        Given a list of courses are existed in DB of "manabie"
        And a list of courses are existed in DB of "above teacher"
        And a list of lesson are existed in DB
        When student retrieves courses
        Then returns "OK" status code
        And returns a list of courses of "manabie & above teacher"

    Scenario: student try to retrieve course with ids
        Given a list of courses are existed in DB of "manabie"
        And a list of courses are existed in DB of "above teacher"
        And a list of lesson are existed in DB
        When student retrieves courses with ids
        Then returns "OK" status code
        And returns a list of courses from requested ids

    Scenario Outline: student try to retrieve courses in current class
        Given a list of courses are existed in DB of "manabie"
        And a list of courses are existed in DB of "above teacher"
        And a list of lesson are existed in DB
        When student retrieves live courses with status "<status>"
        Then returns "OK" status code
        And returns a list of courses of "<list course>"

        Examples:
            | status               | list course |
            | COURSE_STATUS_ACTIVE | live course |
            | COURSE_STATUS_COMPLETED | completed live course |
            | COURSE_STATUS_ON_GOING  |ongoing live course |
