Feature: List student available contents

	Scenario Outline: List student available contents
		Given "student" logins "Learner App"
		And data for list student available contents with "<number_of_books>" book(s)
		When list student available contents
		Then returns "OK" status code
		And verify list contents after list student available contents with "<number_of_books>" book(s)

		Examples:
			| number_of_books |
			| one             |
			| many            |