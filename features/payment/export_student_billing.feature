Feature: Export student billing
	As an HQ manager or admin
	I am able to export student billing data

	Scenario Outline: Admin exports master student billing data archived
		Given the organization "<organization>" has existing student billing data
		When "<signed-in user>" export student billing data
		Then receives "OK" status code
		And the student billing CSV has a correct content

    Examples: 
			| signed-in user | organization |
			| school admin   | -2147483648  | 
			| hq staff       | -2147483648  | 