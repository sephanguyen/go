@blocker
Feature: Import Package Discount Course Mapping

    Background:
		Given there is an existing discount master data with discount tag

    Scenario Outline: Import valid package discount course mapping csv file with correct data
		Given a package discount course mapping payload with "<row-condition>" data
		When "<signed-in user>" imports package discount course mapping
		Then the valid package discount course mapping lines with "<row-condition>" data are imported successfully
		And receives "OK" status code

		Examples:
			| signed-in user | row-condition          |
			| school admin   | all valid rows         |
			| hq staff       | overwrite existing     |

	Scenario Outline: Rollback failed import valid package discount course mapping csv file with incorrect data
		Given a package discount course mapping request payload with incorrect "<row-condition>" data
		When "<signed-in user>" imports package discount course mapping
		Then the import package discount course mapping transaction is rolled back

		Examples:
		| signed-in user | row-condition          |
		| school admin   | empty value row        |
		| hq staff       | invalid value row      |
		| school admin   | valid and invalid rows |

	Scenario Outline: Import invalid package discount course mapping csv file
		Given a package discount course mapping invalid "<invalid format>" request payload
		When "<signed-in user>" imports package discount course mapping
		Then receives "InvalidArgument" status code

		Examples:
		| signed-in user | invalid format                                     |
		| school admin   | no data                                            |
		| hq staff       | header only                                        |
		| school admin   | number of column is not equal 5                    |
		| hq staff       | mismatched number of fields in header and content  |
		| school admin   | wrong package_id column name in header             |
		| hq staff       | wrong course_combination_ids column name in header |
		| hq staff       | wrong discount_tag_id column name in header        |
		| school admin   | wrong is_archived column name in header            |





