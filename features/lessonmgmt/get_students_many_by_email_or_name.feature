Feature: Get students many by email or name

	Background:
		Given user signed in as school admin
		And have some student accounts

	Scenario Outline: Admin get student subscriptions
		When user get students by email or name: "<keyword>"
		Then returns "OK" status code
		And got list students by email or name: "<keyword>"
		Examples:
			| keyword |
			|         |
			| user    |
			| @	      |
