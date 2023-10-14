@major
Feature: Export bank branch
	As an HQ manager or admin
	I am able to export bank branch data

	Scenario Outline: Admin exports master bank branch data archived
		Given the organization "<organization>" has existing "<is-archived>" bank branch data
		When "<signed-in user>" logins to backoffice app
		And the user export bank branch data
		Then receives "OK" status code
		And the bank branch CSV has a correct content

		Examples:
			| signed-in user | organization | is-archived  |
			| school admin   | -2147483630  | archived     |
			| hq staff       | -2147483630  | not archived |

	Scenario Outline: Admin exports master bank branch with no record
		Given the organization "<organization>" has no existing bank branch data
		And "<signed-in user>" logins to backoffice app
		When the user export bank branch data
		Then receives "OK" status code
		And the bank branch CSV only contains the header record

		Examples:
			| signed-in user | organization |
			| school admin   | -2147483629  |
			| hq staff       | -2147483629  |
