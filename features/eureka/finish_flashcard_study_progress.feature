Feature: update The Flashcard study progress
    In order to finish learning an objective or doing an exam
    As a student
    I need to take the quiz

    Background: given a quizet of an learning objective
        Given a quizset with "21" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic

    Scenario Outline: a student finish flashcard study without restart when number of remembered questions aren't equal with number of questions
        Given "student" logins "CMS"
        And user create flashcard study with valid request and limit "20" the first time
        When user finish flashcard study without restart and remembered questions
        Then returns "OK" status code

    Scenario Outline: a student finish flashcard study without restart
        Given "student" logins "CMS"
        And user create flashcard study with valid request and limit "20" the first time
        When user finish flashcard study without restart
        Then returns "OK" status code
        And verify data after finish flashcard without restart

    Scenario Outline: a student finish flashcard study with restart
        Given "student" logins "CMS"
        And user create flashcard study with valid request and limit "20" the first time
        When user finish flashcard study with restart
        Then returns "OK" status code
        And verify data after finish flashcard with restart