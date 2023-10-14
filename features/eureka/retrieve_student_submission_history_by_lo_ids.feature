Feature: Retrieve student submission history by lo_ids
  As a student
  I want to see if I have submitted question for these lo_ids or not
  
    Background: given a quizet of an learning objective
      Given a quiz set with "9" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic

    Scenario: student retrieve student's submission history on a study plan item
      Given "1" students do test of a study plan item
      When student retrieve submission history by lo_ids
      Then returns "OK" status code
