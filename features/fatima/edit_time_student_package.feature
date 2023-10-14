@quarantined
Feature: Edit time student package
  Background:
    Given some package data in db

  Scenario: Unauthenticated user try to edit time student package
    Given an invalid authentication token
    When user edit time a "valid" student package
    Then returns "Unauthenticated" status code

  Scenario Outline: Authenticate user try to edit time student package with invalid case
    Given a signed in "<signed-in user>"
    When user edit time a "<packageId>" student package
    Then returns "<code>" status code

    Examples:
      | signed-in user | packageId | code            |
      | school admin   | empty     | InvalidArgument |
      | school admin   | not exist | NotFound        |

  Scenario Outline: Authenticate user try to edit time student package with valid case
    Given a signed in "<signed-in user>"
    When user edit time a "<packageId>" student package with time from "<startTime>" to "<endTime>"
    Then returns "OK" status code
    And server must store this student package with time from "<startTime>" to "<endTime>"

    Examples:
      | signed-in user | packageId  | startTime               | endTime               |
      | school admin   | valid id   | 2021-02-09T21:46:43Z    | 2031-02-09T21:46:43Z  |

  Scenario Outline: Authenticate user try to edit time student package with valid case
    Given a signed in "<signed-in user>"
    When user edit time a "<packageId>" student package with time from "<startTime>" to "<endTime>" and student package extra
    Then returns "OK" status code
    And server must store this student package with time from "<startTime>" to "<endTime>" and class

    Examples:
      | signed-in user | packageId  | startTime               | endTime               |
      | school admin   | valid id   | 2021-02-09T21:46:43Z    | 2031-02-09T21:46:43Z  |
