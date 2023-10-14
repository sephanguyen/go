Feature: OpenAPI Upsert student by username
  As a school staff
  I need to be able to create a new student by OpenAPI with username

  Background: Sign in with role "staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario: Create students by import OpenAPI file with "<condition>" and available username successfully
    When school admin creates 2 students with "available username" by OpenAPI in folder "students"
    Then students were upserted successfully by OpenAPI

  Scenario Outline: Create students by OpenAPI with "<condition>" unsuccessfully
    When school admin creates 1 students with "<condition>" by OpenAPI in folder "students"
    Then student were created unsuccessfully by OpenAPI with code "<code>" and field "username"

    Examples:
      | condition                              | code  |
      | field username with spaces             | 40001 |
      | field username empty value             | 40001 |
      | field username with special characters | 40004 |
      | without field username                 | 40001 |
      | with existing username                 | 40002 |
      | with existing username and upper case  | 40002 |
