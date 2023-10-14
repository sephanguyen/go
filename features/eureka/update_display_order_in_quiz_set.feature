Feature: Update quiz display order in quiz set
	In user want to change display order in quiz set by change the order between quizzes in the quiz set

	Background: given a quizet of an learning objective
		Given a signed in "school admin"
		And a quizset with "13" quizzes in Learning Objective belonged to a "TOPIC_TYPE_EXAM" topic

	Scenario Outline: authenticate when update quiz display order in quiz set
		Given a signed in "<role>"
		When user change order with "3" times in quiz set
		Then returns "<status code>" status code

		Examples:
			| role           | status code      |
			| school admin   | OK               |
			| admin          | OK               |
			| hq staff       | OK               |
			| student        | PermissionDenied |
			| teacher        | PermissionDenied |
			| parent         | PermissionDenied |
			| center lead    | PermissionDenied |
			| center manager | PermissionDenied |
			| center staff   | PermissionDenied |

	Scenario Outline: update quiz display order between 1 quiz pair
		When user change order with "<num>" times in quiz set
		Then update the order quizzes in quiz set as expected

		Examples:
			| num |
			| 1   |
			| 2   |
			| 3   |
			| 9   |
			| 13  |

	Scenario Outline: move one quiz up/down
		When user move one quiz "<action>" in quiz set
		Then update the order quizzes in quiz set as expected

		Examples:
			| action |
			| up     |
			| down   |
			| none   |