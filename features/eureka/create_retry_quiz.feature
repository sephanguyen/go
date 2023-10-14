Feature: Create retry quiz test

    Background: given a quizset of a learning objective
        Given a quiz set with "10" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic
    
    Scenario: Student do wrong some quizzes and do retry 
        Given a signed in "student"
        And student does the quiz set and wrong some quizzes
        When the student choose option retry quiz
        Then our system have to return the retry quizzes correctly

    Scenario: Student do wrong some quizzes and do retry with quizset containing question group
        Given existing question group
        And user upsert a valid "questionGroup" single quiz
        And a signed in "student"
        And student does the quiz set and wrong some quizzes
        When the student choose option retry quiz
        Then our system have to return the retry quizzes correctly
        And user got quiz test response
