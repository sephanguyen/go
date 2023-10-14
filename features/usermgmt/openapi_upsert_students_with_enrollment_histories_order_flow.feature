@blocker
Feature: Open API Upsert Student With Enrollment Status History In Order Flow

  Background: Sign in with role "staff granted role school admin"
    Given a signed in as "staff granted role school admin" in "kec-demo" organization

  Scenario Outline: Create students by OpenAPI with "<condition>" unsuccessfully in order flow
    When school admin creates 1 students with "<condition>" by OpenAPI in folder "enrollment_status_histories"
    Then students were upserted successfully by OpenAPI

    Examples:
      | condition                                 |
      | enrollment status temporary and potential |
      | enrollment status potential               |
      | enrollment status non-potential           |


  Scenario Outline: Create students by OpenAPI with "<condition>" unsuccessfully in order flow
    When school admin creates 1 students with "<condition>" by OpenAPI in folder "enrollment_status_histories"
    Then student were created unsuccessfully by OpenAPI with code "40004" and field "enrollment_status"

    Examples:
      | condition                   |
      | enrollment status enrolled  |
      | enrollment status graduate  |
      | enrollment status withdraw  |
      | enrollment status temporary |

  Scenario Outline: Update students by OpenAPI with "enrollment status potential" to "<desired-enrollment_status>" unuccessfully in order flow
    Given school admin creates 1 students with "enrollment status potential" by OpenAPI in folder "enrollment_status_histories"
    When school admin updates 1 students with "<desired-enrollment_status>" by OpenAPI
    Then student were updated unsuccessfully by OpenAPI with code "40004" and field "enrollment_status"

    Examples:
      | desired-enrollment_status                           |
      | changing to enrollment status graduate and new date |
      | changing to enrollment status withdraw and new date |
      | changing to enrollment status LOA and new date      |
      | changing to enrollment status enrolled new date     |