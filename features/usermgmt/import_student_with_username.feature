Feature: Import Student with username
  As a school staff
  I need to be able to import the new students with username data

  Background: Sign in with role "staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario: Creating a new student with general information and unique username via Import Student By CSV successfully
    When school admin create 1 students with "with available username" by import in folder "students"
    Then students were upserted successfully by import

  Scenario Outline: Creating a new student with general information and "<condition>" via Import Student By CSV unsuccessfully
    When school admin create 1 students with "<condition>" by import in folder "students"
    Then student were created unsuccessfully by import with code "<code>" and field "username" at row "2"

    Examples:
      | condition                              | code  |
      | field username with multiple spaces    | 40001 |
      | field username with single space       | 40001 |
      | field username empty value             | 40001 |
      | field username with special characters | 40004 |
      | without field username                 | 40001 |
      | with existing username                 | 40002 |
      | with existing username and upper case  | 40002 |

  Scenario Outline: Updating existing students with general information and <update condition> via Import Student By CSV successfully
    Given school admin create 2 students with "with available username" by import in folder "students"
    When school admin update 2 students with "<update condition>" by import
    Then students were upserted successfully by import

    Examples:
      | update condition                                        |
      | keep existing username                                  |
      | editing to another available username                   |
      | editing to another available username with email format |

  Scenario Outline: Updating existing students with general information and <condition> via Import Student By CSV successfully
    Given school admin create 1 students with "all fields" by import in folder "students"
    When school admin update 1 students with "<condition>" by import
    Then student were updated unsuccessfully by import with code "<code>" and field "username" at row "<row>"

    Examples:
      | condition                                   | row | code  |
      | editing to existing username                | 2   | 40002 |
      | editing to existing username and upper case | 2   | 40002 |
      | editing to multiple spaces username         | 2   | 40001 |
      | editing to single space username            | 2   | 40001 |
      | editing to special characters username      | 2   | 40004 |
      | editing to empty username                   | 2   | 40001 |

  Scenario Outline: Update duplicated usernames students with "<condition>" by import unsuccessfully
    Given school admin create 2 students with "all fields" by import in folder "students"
    When school admin update 2 students with "<condition>" by import
    Then student were updated unsuccessfully by import with code "<code>" and field "username" at row "<row>"

    Examples:
      | condition                      | row | code  |
      | editing to duplicated username | 3   | 40003 |
