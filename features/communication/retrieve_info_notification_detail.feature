Feature: Retrieve a notification detail

    user want to read notification detail which is sent to him/her

    @blocker
	Scenario: Retrieve notification detail
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students
        And student logins to Learner App
        And school admin sends some notificationss to a student
        When student retrieve notification detail
        Then returns correct notification detail

    @blocker
	Scenario: Retrieve list of notifications
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students
        And student logins to Learner App
        And school admin sends some notificationss to a student
        When student retrieves list of notifications
        Then returns correct list of notifications

    @blocker
	Scenario: Count user notification by status
        Given a new "staff granted role school admin" and granted organization location logged in Back Office of a new organization with some exist locations
        And school admin creates "1" students
        And student logins to Learner App
        And school admin sends some notificationss to a student
        And student reads some notifications
        When student counts number of read notification
        Then returns correct number of read notification
