@blocker
Feature: Reallocate student enrollment status

  Background: Sign in with role "staff granted role school admin"
    Given a signed in "staff granted role school admin"

  Scenario: Reallocate student enrollment status successfully
    Given school admin create a student with "general info" and "mandatory fields" by GRPC
    When school admin reallocate the student's enrollment status with "temporary status" data
    Then student's enrollment status was reallocated successfully

  Scenario Outline: Reallocate student enrollment status failed due to invalid data
    Given school admin create a student with "general info" and "mandatory fields" by GRPC
    When school admin reallocate the student's enrollment status with "<invalid data type>" data
    Then school admin can not reallocate the student's enrollment status

    Examples:
      | invalid data type                  |
      | student does not existed           |
      | location does not existed          |
      | enrollment status is not temporary |
      | start date after end date          |
