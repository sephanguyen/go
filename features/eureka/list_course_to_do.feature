Feature: List course-study plan's to do items

	List course-study plan's to do items with general statistic

	Scenario Outline: List course-study plan's to do items
		# take the advantage of this step which include:
		# 	- create a course
		# 	- create study plan with some assignments
		#	- assign that study plan to course
		# 	- some students are assigned some valid study plans
		Given some students are assigned some valid study plans
		And student submit their "existed" content assignment "<times>" times

		When list course-study plan's to do items
		Then returns list of to do items with correct statistic infor

		Examples:
			| times    |
			| multiple |
			| single   |

	Scenario Outline: List course-study plan's to do items with failed case
		Given some students are assigned some valid study plans
		And student submit their "existed" content assignment "<times>" times

		When list course-study plan's to do items with "<failed case>"
		Then returns empty list of to do items

		Examples:
			| times    | failed case            |
			| multiple | empty study plan       |
			| single   | not existed study plan |

	Scenario: List course-study plan's to do items with items were deleted
		Given some students are assigned some valid study plans
		And delete study plan items by study plans
		When list course-study plan's to do items
		Then returns empty list of to do items
