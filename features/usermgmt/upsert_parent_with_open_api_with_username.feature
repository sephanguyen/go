Feature: Upsert parent with open api with username
  As a HQ staff
  I want to be able to upsert a parent with username

  Background: Sign in with role"staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario Outline: Create a parent with "<condition>" successfully
    When school admin creates parents "valid" with "<condition>" by OpenAPI
    Then parents were created by OpenAPI successfully

    Examples:
      | condition                            |
      | available username                   |
      | available username with email format |


  Scenario Outline: Create a parent with invalid field unsuccessfully with "<condition>"
    When school admin creates parents "invalid" with "<condition>" by OpenAPI
    Then parents were created by OpenAPI unsuccessfully with "<code>" code and "username" field

    Examples:
      | condition                                  | code  |
      | username was used by other                 | 40002 |
      | username was used by other with upper case | 40002 |
      | empty username                             | 40001 |
      | username has special characters            | 40004 |
