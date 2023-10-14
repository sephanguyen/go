Feature: Retrieve total quiz of list of learning objective
	As a student, in every leraning objective i want to know how many quiz which i have to do

	Background:
		Given a learning objective belonged to a "TOPIC_TYPE_EXAM" topic has quizset with 3 quizzes
		And a learning objective belonged to a "TOPIC_TYPE_EXAM" topic has quizset with 9 quizzes
		And a learning objective belonged to a "TOPIC_TYPE_EXAM" topic has quizset with 11 quizzes
		And a learning objective belonged to a "TOPIC_TYPE_EXAM" topic has quizset with 17 quizzes
		And a learning objective belonged to a "TOPIC_TYPE_EXAM" topic has quizset with 19 quizzes
		And a learning objective belonged to a "TOPIC_TYPE_EXAM" topic has quizset with 0 quizzes
		And a learning objective belonged to a "TOPIC_TYPE_EXAM" topic has no quizset

    Scenario: unauthenticated user try to create the quiz test
        Given an invalid authentication token
		When user get total quiz of lo "1, 2, 3"
        Then returns "Unauthenticated" status code

    Scenario:  get total quiz of los missing lo ids
		Given a signed in "student"
		When user get total quiz of lo without lo ids
		Then returns "InvalidArgument" status code

    Scenario:  get total quiz of los
		Given a signed in "student"
		When user get total quiz of lo "1, 2, 3"
		Then returns "OK" status code
		And total quiz set is "3, 9, 11"
		
    Scenario: student get total quiz of los
		Given a signed in "student"
		When user get total quiz of lo "1, 2, 3"
		Then returns "OK" status code
		And total quiz set is "3, 9, 11"

    Scenario: student get total quiz of los
		Given a signed in "student"
		When user get total quiz of lo "3, 4, 5"
		Then returns "OK" status code
		And total quiz set is "11, 17, 19"

    Scenario: student get total quiz of los
		Given a signed in "student"
		When user get total quiz of lo "1, 3, 5"
		Then returns "OK" status code
		And total quiz set is "3, 11, 19"

    Scenario: student get total quiz of los, but los have no quiz
		Given a signed in "student"
		When user get total quiz of lo "6, 1, 7"
		Then returns "OK" status code
		And total quiz set is "0, 3, 0"

    Scenario: student get total quiz of los, but some los have no quiz
		Given a signed in "student"
		When user get total quiz of lo "1, 2, 6"
		Then returns "OK" status code
		And total quiz set is "3, 9, 0"