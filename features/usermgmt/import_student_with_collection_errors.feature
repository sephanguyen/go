@blocker
Feature: Import Student wtih collection errors

  Background: Sign in with role "staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario: Create students by import csv with "all rows are failed"
    When school admin create 10 students with "all rows are failed" by import in folder "students"
    Then students were imported with failed 10 rows and successful 0 rows
      | condition                             | field              | code  | at_row |
      | empty first_name                      | first_name         | 40001 | 2      |
      | empty last_name                       | last_name          | 40001 | 3      |
      | non exsting grade                     | grade              | 40400 | 4      |
      | worng format email                    | email              | 40004 | 5      |
      | worng format birthday                 | birthday           | 40004 | 6      |
      | non exsting prefecture                | prefecture         | 40400 | 7      |
      | school_course is not mapped to school | school_course      | 40004 | 8      |
      | empty location                        | location           | 40004 | 9      |
      | non existing location                 | location           | 40400 | 10     |
      | contact_preference is out if range    | contact_preference | 40004 | 11     |

  Scenario: Create students by import csv with "some rows are successful and failed"
    When school admin create 10 students with "some rows are successful and failed" by import in folder "students"
    Then students were imported with failed 5 rows and successful 5 rows
      | condition                          | field              | code  | at_row |
      | empty last_name                    | last_name          | 40001 | 3      |
      | worng format email                 | email              | 40004 | 5      |
      | non exsting prefecture             | prefecture         | 40400 | 7      |
      | empty location                     | location           | 40004 | 9      |
      | contact_preference is out if range | contact_preference | 40004 | 11     |

  Scenario: Create students by import csv with "all rows are successful"
    When school admin create 10 students with "all rows are successful" by import in folder "students"
    Then students were upserted successfully by import


  Scenario: Update students by import csv with "editing to empty first_name"
    Given school admin create 10 students with "all rows are successful" by import in folder "students"
    When school admin update 10 students with "editing to empty first_name" by import
    Then students were imported with failed 10 rows and successful 0 rows
      | condition        | field      | code  | at_row |
      | empty first_name | first_name | 40001 | 2      |
      | empty first_name | first_name | 40001 | 3      |
      | empty first_name | first_name | 40001 | 4      |
      | empty first_name | first_name | 40001 | 5      |
      | empty first_name | first_name | 40001 | 6      |
      | empty first_name | first_name | 40001 | 7      |
      | empty first_name | first_name | 40001 | 8      |
      | empty first_name | first_name | 40001 | 9      |
      | empty first_name | first_name | 40001 | 10     |
      | empty first_name | first_name | 40001 | 11     |

  Scenario: Update students by import csv with "editing to empty external_user_id"
    Given school admin create 10 students with "all rows are successful" by import in folder "students"
    When school admin update 5 students with "editing to empty external_user_id" by import
    Then students were imported with failed 5 rows and successful 5 rows
      | condition              | field            | code  | at_row |
      | empty external_user_id | external_user_id | 40008 | 2      |
      | empty external_user_id | external_user_id | 40008 | 3      |
      | empty external_user_id | external_user_id | 40008 | 4      |
      | empty external_user_id | external_user_id | 40008 | 5      |
      | empty external_user_id | external_user_id | 40008 | 6      |