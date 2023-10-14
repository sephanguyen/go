@blocker
Feature: Bulk Update Students

    Background: Sign in with role "staff granted role school admin"
        Given a signed in "staff granted role school admin"

    Scenario Outline: Update a student with "<condition>" by import successfully
        Given school admin create 2 students with "<row condition>" by import in folder "students"
        When school admin update 2 students with "<condition>" by import
        Then students were upserted successfully by import

        Examples:
            | row condition                      | condition                         |
            | all fields                         | empty enrollment_status_histories |
            | all fields                         | edit first_name                   |
            | all fields                         | edit last_name                    |
            | all fields                         | edit first_name_phonetic          |
            | all fields                         | edit last_name_phonetic           |
            | all fields                         | edit birthday                     |
            | all fields                         | edit gender                       |
            | all fields                         | edit grade                        |
            | all fields                         | edit external_user_id with spaces |
            | mandatory fields                   | edit to have external_user_id     |
            | empty external user id with spaces | edit to have external_user_id     |
    # | edit email                        | TODO: now we skip email @danh will update them later

    Scenario Outline: Update a student with "<condition>" by import unsuccessfully
        Given school admin create 2 students with "all fields" by import in folder "students"
        When school admin update 1 students with "<condition>" by import
        Then student were updated unsuccessfully by import with code "<code>" and field "<field>" at row "<row>"

        Examples:
            | condition                                     | field            | row | code  |
            | editing to non-existing user_id               | user_id          | 2   | 40400 |
            | editing to non-existing grade                 | grade            | 2   | 40400 |
            | editing to wrong format birthday              | birthday         | 2   | 40004 |
            | editing to out of range gender                | gender           | 2   | 40004 |
            | editing to text gender                        | gender           | 2   | 40004 |
            | editing to non-existing school                | school           | 2   | 40400 |
            | editing to non-existing school_course         | school_course    | 2   | 40400 |
            | editing to empty first_name                   | first_name       | 2   | 40001 |
            | editing to empty last_name                    | last_name        | 2   | 40001 |
            | editing to empty external_user_id             | external_user_id | 2   | 40008 |
            | editing to empty external_user_id with spaces | external_user_id | 2   | 40008 |

    Scenario Outline: Update duplicated students with "<condition>" by import unsuccessfully
        Given school admin create 2 students with "all fields" by import in folder "students"
        When school admin update 2 students with "<condition>" by import
        Then student were updated unsuccessfully by import with code "<code>" and field "<field>" at row "<row>"

        Examples:
            | condition                              | field            | row | code  |
            | editing to duplicated user_id          | user_id          | 3   | 40003 |
            | editing to duplicated external_user_id | external_user_id | 3   | 40003 |
    # | editing to duplicated email            | email            | 3   | 40003 | TODO: now we skip email @danh will update them later

    Scenario: Update a student with "editing to existing external_user_id" by import unsuccessfully
        Given school admin create 2 students with "mandatory fields" by import in folder "students"
        When school admin update 1 students with "editing to existing external_user_id" by import
        Then student were updated unsuccessfully by import with code "40002" and field "external_user_id" at row "2"
