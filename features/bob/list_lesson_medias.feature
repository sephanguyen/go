@quarantined
Feature: List lesson medias
	In order to find medias of a lesson

	Background: student upsert media
        Given a list of media which attached to a lesson

	Scenario: admin try to list lesson medias
        Given "staff granted role school admin" signin system
        When user get lesson medias and returns "PermissionDenied" status code

	Scenario: student try to list lesson medias
        Given "student" signin system
        When user get lesson medias
        Then returns "OK" status code
        And the list of media match with response medias

	Scenario: teacher try to list lesson medias
        Given "staff granted role teacher" signin system
        When user get lesson medias
        Then returns "OK" status code
        And the list of media match with response medias

	Scenario: user get media of non existed lesson
        Given "staff granted role teacher" signin system
        When user get media of non existed lesson
        Then returns "OK" status code
        And empty media result