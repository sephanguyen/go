Feature: Import Lessons
  Background:
    Given user signed in as school admin 
    And have some centers
    And "disable" Unleash feature with feature name "Lesson_LessonManagement_BackOffice_ImportLessonByCSVV2"

  Scenario Outline: Import valid csv file
    Given "<signed-in user>" signin system
    And a valid lessons payload
    When importing lessons
    Then returns "OK" status code
    And the valid lessons lines are imported successfully
    
    Examples:
      | signed-in user                  |
      | school admin                    |
      | teacher                         |
      | staff granted role school admin |

  Scenario Outline: Import invalid csv file
    Given "<signed-in user>" signin system
    And an invalid lessons "<invalid format>" request payload
    When importing lessons
    Then returns "OK" status code
    And the invalid lessons must returned with error
    
    Examples:
      | signed-in user                  | invalid format                                    |
      | staff granted role school admin | no data                                           |
      | staff granted role school admin | header only                                       |
      | school admin                    | number of column is not equal 4                   |
      | school admin                    | mismatched number of fields in header and content |
      | staff granted role school admin | wrong id column name in header                    |
      | staff granted role school admin | wrong name column name in header                  |
      | teacher                         | mismatched valid and invalid rows                 |
      | teacher                         | invalid partner_internal_id                       |
      | teacher                         | invalid date time format                          |
      | staff granted role school admin | start time > end time                             |
      | staff granted role school admin | start date <> end date                            |
      | school admin                    | invalid teaching_method                           |
      | staff granted role school admin | missing value in madatory column                  |
