@blocker
Feature: OpenAPI Upsert student
  As a school staff
  I need to be able to create a new student

  Background: Sign in with role "staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario Outline: Create students by import csv file with "<condition>" successfully
    When school admin creates 2 students with "<condition>" by OpenAPI in folder "students"
    Then students were upserted successfully by OpenAPI

    Examples:
      | condition                    |
      | mandatory fields             |
      | all fields                   |
      | empty phonetic name          |
      | external user id with spaces |

  Scenario Outline: Create students by openAPI with "<condition>" unsuccessfully
    When school admin creates 1 students with "<condition>" by OpenAPI in folder "students"
    Then student were created unsuccessfully by OpenAPI with code "<code>" and field "<field>"

    Examples:
      | condition                                                  | field             | code  |
      | out of gender range                                        | gender            | 40004 |
      | non existing grade                                         | grade             | 40400 |
      | wrong format email                                         | email             | 40004 |
      | non existing tags                                          | student_tag       | 40400 |
      | non existing enrollment_status_histories location          | location          | 40400 |
      | out of range enrollment_status_histories enrollment_status | enrollment_status | 40004 |
      | missing external_user_id                                   | external_user_id  | 40001 |
      | empty external user id with spaces                         | external_user_id  | 40001 |
      | missing first_name                                         | first_name        | 40001 |
      | missing last_name                                          | last_name         | 40001 |
      | missing email                                              | email             | 40001 |
      | missing grade                                              | grade             | 40001 |
      | missing enrollment_status_histories                        | enrollment_status | 40001 |
      | missing enrollment_status_histories location               | location          | 40004 |
      | missing enrollment_status_histories enrollment_status      | enrollment_status | 40004 |

  Scenario Outline: Create a student with student info and address "<condition>"
    When school admin creates 1 students with "<condition>" by OpenAPI in folder "address"
    Then students were upserted successfully by OpenAPI

    Examples:
      | condition          |
      | city only          |
      | postal code only   |
      | prefecture only    |
      | first street only  |
      | second street only |

  Scenario Outline: Create a student with student info and phone number "<condition>"
    When school admin creates 1 students with "<condition>" by OpenAPI in folder "phone_number"
    Then students were upserted successfully by OpenAPI

    Examples:
      | condition                                              |
      | contact preference only                                |
      | phone number only                                      |
      | home phone number only                                 |
      | both phone number and home phone number is empty value |
      | phone number is empty value                            |


  Scenario Outline: Create a student with student info and school histories "<condition>"
    When school admin creates 1 students with "<condition>" by OpenAPI in folder "school_history"
    Then students were upserted successfully by OpenAPI

    Examples:
      | condition                                                                |
      | only school                                                              |
      | school and school_course                                                 |
      | school and school_course and start_date                                  |
      | there is not current_school                                              |
      | 3 school 2 school_course 1 empty school_course empty start_date end_date |
      | 2 school 2 school_course 1 start_date 1 end_date                         |


  Scenario Outline: Create students invalid school histories by OpenAPI with "<row condition>" unsuccessfully
    When school admin creates 1 students with "<row condition>" by OpenAPI in folder "school_history"
    Then student were created unsuccessfully by OpenAPI with code "<code>" and field "<field>"

    Examples:
      | row condition                          | field         | code  |
      | empty school but there is other fields | school        | 40004 |
      | school_course is not mapped to school  | school_course | 40004 |
      | non existing school                    | school        | 40400 |
      | non existing school_course             | school_course | 40400 |
      | start_date after end_date              | start_date    | 40004 |

  Scenario Outline: Upsert a student with "<condition>" by open api successfully
    Given school admin creates 1 students with "all fields" by OpenAPI in folder "students"
    When school admin updates 1 students with "<condition>" by OpenAPI
    Then students were upserted successfully by OpenAPI

    Examples:
      | condition                                                        |
      | empty enrollment_status_histories                                |
      | edit first_name                                                  |
      | edit last_name                                                   |
      | edit first_name_phonetic                                         |
      | edit last_name_phonetic                                          |
      | edit birthday                                                    |
      | edit gender                                                      |
      | edit grade                                                       |
      | edit with empty locations, other info still remains              |
      | edit with external_user_id with spaces, other info still remains |
  # | edit email                           | TODO: @shanenoi will uncomment this later, because OpenAPI wont support update email (LT-37895)

  Scenario Outline: Upsert a student with "<condition>" by OpenAPI unsuccessfully
    Given school admin creates 1 students with "all fields" by OpenAPI in folder "students"
    When school admin updates 1 students with "<condition>" by OpenAPI
    Then student were updated unsuccessfully by OpenAPI with code "<code>" and field "<field>"

    Examples:
      | condition                                      | field            | code  |
      | editing to non-existing external_user_id       | email            | 40002 |
      | editing to non-existing grade                  | grade            | 40400 |
      | editing to out of range gender                 | gender           | 40004 |
      | editing to non-existing school                 | school           | 40400 |
      | editing to non-existing school_course          | school_course    | 40400 |
      | editing to empty first_name                    | first_name       | 40001 |
      | editing to empty last_name                     | last_name        | 40001 |
      | editing to empty external_user_id              | external_user_id | 40001 |
      | editing to empty external_user_id with spaces  | external_user_id | 40001 |
      | editing to external_user_id was used by parent | external_user_id | 40002 |

  Scenario: Upsert a student with "editing to duplicated external_user_id" by OpenAPI unsuccessfully
    Given school admin creates 2 students with "all fields" by OpenAPI in folder "students"
    When school admin updates 2 students with "editing to duplicated external_user_id" by OpenAPI
    Then student were updated unsuccessfully by OpenAPI with code "40003" and field "user_id"
# TODO: the error should be external_user_id instead of user_id. Will fix this later
