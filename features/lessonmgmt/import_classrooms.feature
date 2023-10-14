Feature: Import Classrooms
  Background:
    Given user signed in as school admin 
    And have some centers

  Scenario Outline: Import valid csv file
    Given "<signed-in user>" signin system
    And a valid classrooms payload
    When importing classrooms
    Then returns "OK" status code
    And the valid classroom lines are imported successfully
    
    Examples:
      | signed-in user                  |
      | school admin                    |
      | teacher                         |
      | staff granted role school admin |
  
  Scenario Outline: Import invalid csv file
    Given "<signed-in user>" signin system
    And an invalid classrooms "<invalid format>" request payload
    When importing classrooms
    Then returns "OK" status code
    And the invalid classrooms must returned with error
    
    Examples:
      | signed-in user                  | invalid format                                    |
      | staff granted role school admin | no data                                           |
      | staff granted role school admin | header only                                       |
      | school admin                    | mismatched number of fields in header and content |
      | staff granted role school admin | wrong id column name in header                    |
      | staff granted role school admin | wrong name column name in header                  |
      | teacher                         | mismatched valid and invalid rows                 |
      | teacher                         | invalid location_id                               |
      | staff granted role school admin | missing value in madatory column                  |
