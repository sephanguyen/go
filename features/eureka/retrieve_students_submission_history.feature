Feature: Retrieve Submission History
	In order to review my student submission
	As a teacher
	I need to get list of all my student's submission on a study plan item

    Background: given a quizet of an learning objective
		Given a study plan item is learning objective belonged to a "TOPIC_TYPE_EXAM" topic which has quizset with "9" quizzes

    Scenario Outline: student try to retrieve student's submission history on a study plan item
		Given "<num>" students do test of a study plan item
		When student retrieve all student's submission history in that study plan item
		Then returns "PermissionDenied" status code

		Examples:
			| num  |
			|  2   |
			|  1   |
			|  7   |
			|  5   |
			|  11  |

    Scenario Outline: teacher try to retrieve student's submission history on a study plan item
		Given "<num>" students do test of a study plan item
		When teacher retrieve all student's submission history in that study plan item
		Then returns "OK" status code
			And each item have returned addition fields for flashcard
			And the ordered of logs must be correct
			And show correct logs info
			And get "<num>" student's submission history

		Examples:
			| num  |
			|  1   |
			|  2   |
			|  7   |
			|  5   |
			|  11  |

	Scenario Outline: teacher try to retrieve student's submission history on a study plan item that quizset having question groups
		Given existing question group
			And user upsert a valid "questionGroup" single quiz
			And user upsert a valid "questionGroup" single quiz
			And "<num>" students do test of a study plan item
		When teacher retrieve all student's submission history in that study plan item
		Then returns "OK" status code
			And each item have returned addition fields for flashcard
			And the ordered of logs must be correct
			And show correct logs info
			And get "<num>" student's submission history

		Examples:
			| num  |
			|  1   |
			|  2   |
			|  7   |
			|  5   |
			|  11  |

    Scenario Outline: teacher try to retrieve student's submission history on a study plan item. But some student drop the quiz test
      	Given "<num>" students do test of a study plan item
			And "<num_not_finish>" students didn't finish the test
      	When teacher retrieve all student's submission history in that study plan item
			And each item have returned addition fields for flashcard
			Then returns "OK" status code
			And the ordered of logs must be correct
			And show correct logs info
			And get "<num_total>" student's submission history

		Examples:
			| num  | num_not_finish | num_total |
			|  1   | 1              | 2         |
			|  2   | 2              | 4         |
			|  7   | 3              | 10        |
			|  5   | 4              | 9         |
			|  11  | 5              | 16        |

	Scenario Outline: teacher try to retrieve student's submission history on a study plan item that quizset having question groups. But some student drop the quiz test
      	Given existing question group
			And user upsert a valid "questionGroup" single quiz
			And user upsert a valid "questionGroup" single quiz
			And "<num>" students do test of a study plan item
			And "<num_not_finish>" students didn't finish the test
      	When teacher retrieve all student's submission history in that study plan item
			And each item have returned addition fields for flashcard
			Then returns "OK" status code
			And the ordered of logs must be correct
			And show correct logs info
			And get "<num_total>" student's submission history

		Examples:
			| num  | num_not_finish | num_total |
			|  1   | 1              | 2         |
			|  2   | 2              | 4         |
			|  7   | 3              | 10        |
			|  5   | 4              | 9         |
			|  11  | 5              | 16        |