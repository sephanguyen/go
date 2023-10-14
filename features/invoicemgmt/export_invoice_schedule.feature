@major
Feature: Export invoice schedule
	As an HQ manager or admin
	I am able to export invoice schedule data

	Scenario Outline: Admin exports master invoice schedule data
		Given the organization "<organization>" has existing "<file-content-type>" import invoice schedules in "<timezone>"
		And "<signed-in user>" logins to backoffice app
		When admin export the invoice schedule data
		Then receives "OK" status code
		And the invoice schedule CSV has a correct content with invoice date in default timezone "COUNTRY_JP"

		Examples:
			| signed-in user | organization | timezone   | file-content-type    |
			| school admin   | -2147483630  | COUNTRY_JP | multiple-valid-dates |
			| school admin   | -2147483630  | COUNTRY_JP | single-valid-date    |
			| hq staff       | -2147483630  | COUNTRY_VN | single-valid-date    |

	Scenario Outline: Admin exports master invoice schedule with no record
		Given the organization "<organization>" has no existing invoice schedule
		And "<signed-in user>" logins to backoffice app
		When admin export the invoice schedule data
		Then receives "OK" status code
		And the invoice schedule CSV only contains the header record

		Examples:
			| signed-in user | organization |
			| school admin   | -2147483629  |
			| hq staff       | -2147483629  |
