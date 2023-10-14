@blocker
Feature: Upsert student
  As a school staff
  I need to be able to upsert a new student

  Background: Sign in with role "staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario Outline: Create a student with only student info with "<condition>"
    When school admin create a student with "general info" and "<condition>" by GRPC
    Then students were upserted successfully by GRPC

    Examples:
      | condition                    |
      | mandatory fields             |
      | all fields                   |
      | empty phonetic name          |
      | external user id with spaces |

  Scenario Outline: Create a student with student info and address with "<condition>"
    When school admin create a student with "address" and "<condition>" by GRPC
    Then students were upserted successfully by GRPC

    Examples:
      | condition          |
      | city only          |
      | postal code only   |
      | prefecture only    |
      | first street only  |
      | second street only |

  Scenario Outline: Create a student with student info and phone number with "<condition>"
    When school admin create a student with "student phone number" and "<condition>" by GRPC
    Then students were upserted successfully by GRPC

    Examples:
      | condition                      |
      | student phone number only      |
      | student home phone number only |

  Scenario: Create a student with student info and school histories
    When school admin create a student with "school history" and "school only" by GRPC
    Then students were upserted successfully by GRPC

  Scenario: Create a student with student info and "existing external_user_id"
    When school admin create a student with "general info" and "duplicated external_user_id" by GRPC
    Then students were upserted unsuccessfully by GRPC with "40002" code and "external_user_id" field

  Scenario: Update a student with student info and "unable edit external_user_id"
    Given school admin create a student with "general info" and "all fields" by GRPC
    When school admin update a student with "edit external_user_id"
    Then students were upserted unsuccessfully by GRPC with "40008" code and "external_user_id" field

  Scenario Outline: Update a student with student info and "<update enrollment status condition>"
    Given school admin create a student with "enrollment status history" and "<init enrollment status condition>" by GRPC
    When school admin update a student with "<update enrollment status condition>"
    Then students were upserted successfully by GRPC

    Examples:
      | init enrollment status condition | update enrollment status condition                  |
      | potential and temporary status   | update end-date temporary enrollment status history |
  # | potential and enrolled status    | update potential status to temporary status         | TODO: need to fix, this case shoud be failed

  Scenario Outline: Update a student with "<condition>" by grpc successfully
    Given school admin create a student with "general info" and "all fields" by GRPC
    When school admin update a student with "<condition>"
    Then students were upserted successfully by GRPC

    Examples:
      | condition                                       |
      | editing to empty enrollment_status and location |
      | adding one more enrollment_status and location |


  Scenario Outline: Upsert a student with "<condition>" by GRPC unsuccessfully
    Given school admin create a student with "general info" and "all fields" by GRPC
    When school admin update a student with "<condition>"
    Then students were upserted unsuccessfully by GRPC with "<code>" code and "<field>" field

    Examples:
      | condition                             | field         | code  |
      | editing to non-existing grade         | grade         | 40400 |
      | editing to non-existing school        | school        | 40400 |
      | editing to non-existing school_course | school_course | 40400 |
      | editing to empty first_name           | first_name    | 40001 |
      | editing to empty last_name            | last_name     | 40001 |