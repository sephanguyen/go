Feature: Upsert student
  As a school staff
  I need to be able to upsert a new student with username data

  Background: Sign in with role "staff granted role school admin" and enable new username feature
    Given a signed in as "staff granted role school admin" in "manabie" organization


  Scenario: Creating a new student with general information and unique username via GRPC
    When school admin create a student with "general info" and "available username" by GRPC
    Then students were upserted successfully by GRPC


  Scenario Outline: Create a student unsuccessfully with general info and <condition> condition via GRPC
    When school admin create a student with "general info" and "<condition>" by GRPC
    Then students were upserted unsuccessfully by GRPC with "<code>" code and "username" field

    Examples:
      | condition                                  | code  |
      | username was used by other                 | 40002 |
      | username was used by other with upper case | 40002 |
      | empty username                             | 40001 |
      | username has special characters            | 40004 |


  Scenario: Updating a existed student with general information and unique username via GRPC
    Given school admin create a student with "general info" and "available username" by GRPC
    When school admin update a student with "another available username"
    Then students were upserted successfully by GRPC


  Scenario Outline: Update a student unsuccessfully with general info and <condition> condition via GRPC
    Given school admin create a student with "general info" and "available username" by GRPC
    When school admin update a student with "<condition>"
    Then students were upserted unsuccessfully by GRPC with "<code>" code and "username" field

    Examples:
      | condition                                  | code  |
      | username was used by other                 | 40002 |
      | username was used by other with upper case | 40002 |
      | empty username                             | 40001 |
      | username has special characters            | 40004 |
