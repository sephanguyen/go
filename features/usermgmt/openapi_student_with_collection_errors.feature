@blocker
Feature: Upsert students by OpenAPI wtih collection errors

  Background: Sign in with role "staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario: Create students by OpenAPI with "all rows are failed"
    When school admin creates 10 students with "all rows are failed" by OpenAPI in folder "students"
    Then students were upserted by OpenAPI with failed 10 rows and successful 0 rows
      | condition                             | field              | code  | index |
      | empty first_name                      | first_name         | 40001 | 0     |
      | empty last_name                       | last_name          | 40001 | 1     |
      | non exsting grade                     | grade              | 40400 | 2     |
      | worng format email                    | email              | 40004 | 3     |
      | duplicated home_phone_number          | home_phone_number  | 40003 | 4     |
      | non exsting prefecture                | prefecture         | 40400 | 5     |
      | school_course is not mapped to school | school_course      | 40004 | 6     |
      | empty location                        | location           | 40004 | 7     |
      | non existing location                 | location           | 40400 | 8     |
      | contact_preference is out if range    | contact_preference | 40004 | 9     |

  Scenario: Create students by OpenAPI with "some rows are successful and failed"
    When school admin creates 10 students with "some rows are successful and failed" by OpenAPI in folder "students"
    Then students were upserted by OpenAPI with failed 5 rows and successful 5 rows
      | condition                          | field              | code  | index |
      | empty last_name                    | last_name          | 40001 | 1     |
      | worng format email                 | email              | 40004 | 3     |
      | non exsting prefecture             | prefecture         | 40400 | 5     |
      | empty location                     | location           | 40004 | 7     |
      | contact_preference is out if range | contact_preference | 40004 | 9     |

  Scenario: Create students by OpenAPI with "all rows are successful"
    When school admin creates 10 students with "all rows are successful" by OpenAPI in folder "students"
    Then students were upserted successfully by OpenAPI


  Scenario: Update students by OpenAPI with "editing to empty first_name"
    Given school admin creates 10 students with "all rows are successful" by OpenAPI in folder "students"
    When school admin updates 10 students with "editing to empty first_name" by OpenAPI
    Then students were upserted by OpenAPI with failed 10 rows and successful 0 rows
      | condition        | field      | code  | index |
      | empty first_name | first_name | 40001 | 0     |
      | empty first_name | first_name | 40001 | 1     |
      | empty first_name | first_name | 40001 | 2     |
      | empty first_name | first_name | 40001 | 3     |
      | empty first_name | first_name | 40001 | 4     |
      | empty first_name | first_name | 40001 | 5     |
      | empty first_name | first_name | 40001 | 6     |
      | empty first_name | first_name | 40001 | 7     |
      | empty first_name | first_name | 40001 | 8     |
      | empty first_name | first_name | 40001 | 9     |

  Scenario: Update students by OpenAPI with "editing to empty external_user_id"
    Given school admin creates 10 students with "all rows are successful" by OpenAPI in folder "students"
    When school admin updates 5 students with "editing to empty external_user_id" by OpenAPI
    Then students were upserted by OpenAPI with failed 5 rows and successful 5 rows
      | condition              | field            | code  | index |
      | empty external_user_id | external_user_id | 40001 | 0     |
      | empty external_user_id | external_user_id | 40001 | 1     |
      | empty external_user_id | external_user_id | 40001 | 2     |
      | empty external_user_id | external_user_id | 40001 | 3     |
      | empty external_user_id | external_user_id | 40001 | 4     |