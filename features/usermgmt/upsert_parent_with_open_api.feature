@blocker
Feature: Upsert parent with open api
  As a HQ staff
  I want to be able to upsert a parent

  Background: Sign in with role"staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario Outline: Create a parent with "<condition>" successfully
    When school admin creates parents "valid" with "<condition>" by OpenAPI
    Then parents were created by OpenAPI successfully

    Examples:
      | condition                    |
      | mandatory fields             |
      | with tags                    |
      | with user phone numbers      |
      | existed user in database     |
      | create multiple parents      |
      | both create and update       |
      | external user id with spaces |

  Scenario Outline: Create a parent with invalid field unsuccessfully with "<condition>"
    When school admin creates parents "invalid" with "<condition>" by OpenAPI
    Then parents were created by OpenAPI unsuccessfully with "<code>" code and "<field>" field

    Examples:
      | condition                            | code  | field            |
      | invalid tags                         | 40004 | parent_tag       |
      | invalid email                        | 40004 | email            |
      | missing email                        | 40001 | email            |
      | missing first_name                   | 40001 | first_name       |
      | missing last_name                    | 40001 | last_name        |
      | missing external_user_id             | 40001 | external_user_id |
      | missing children                     | 40001 | children         |
      | children empty                       | 40001 | children         |
      | invalid children email               | 40004 | student_email    |
      | children email is empty              | 40001 | student_email    |
      | children email is null               | 40001 | student_email    |
      | children relationship is missing     | 40001 | relationship     |
      | children relationship is invalid     | 40004 | relationship     |
      | external_user_id was used by student | 40002 | external_user_id |
