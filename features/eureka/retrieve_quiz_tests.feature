Feature: Retrieve quiz tests
  I'm a teacher, in a study plan item i want to get all students's tests info

  Background: given a quizet of an learning objective
    Given a quiz set with "9" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic

  Scenario: students take a quiz test successfully
    Given "3" students do test of a study plan item
    When teacher get quiz test of a study plan item
    Then returns "OK" status code
    And 3 quiz tests infor
    And compare quiz tests list with bob service

  Scenario: students take a quiz test successfully
    Given "10" students do test of a study plan item
    When teacher get quiz test of a study plan item
    Then returns "OK" status code
    And 10 quiz tests infor

  Scenario: students take a quiz test successfully
    Given "50" students do test of a study plan item
    When teacher get quiz test of a study plan item
    Then returns "OK" status code
    And 50 quiz tests infor

  Scenario: students missing study plan item id
    Given "3" students do test of a study plan item
    When teacher get quiz test without study plan item id
    Then returns "InvalidArgument" status code