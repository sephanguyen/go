@quarantined
Feature: Join live token
  In order for user to see live stream
  As a user
  I need to join lesson

  Background:
    Given "staff granted role school admin" signin system
    And a random number
    And a school name "S1", country "COUNTRY_VN", city "Hồ Chí Minh", district "2"
    And a school name "S2", country "COUNTRY_VN", city "Hồ Chí Minh", district "3"
    And admin inserts schools

    Given "teacher" signin system

  Scenario: student join lesson
    Given "student" signin system
    And a list of courses are existed in DB of "above teacher"
    And a student with valid lesson
    When student join lesson
    Then returns "OK" status code
    And student must receive lesson room id

  Scenario: student join lesson without permission
    Given "student" signin system
    And a list of courses are existed in DB of "above teacher"
    When student join lesson
    Then returns "PermissionDenied" status code

  Scenario: student join lesson expecting rtm token
    Given "student" signin system
    And a list of courses are existed in DB of "above teacher"
    And a student with valid lesson
    When student join lesson with v1 API
    Then returns "OK" status code
    And student must receive lesson room id and tokens

  Scenario: student join lesson which has no room id
    Given "student" signin system
    And a list of courses are existed in DB of "above teacher"
    And a student with valid lesson which has no room id
    When student join lesson
    Then returns "OK" status code
    And student must receive lesson room id

    When student join lesson again
    Then returns "OK" status code
    And student must receive lesson room id same above