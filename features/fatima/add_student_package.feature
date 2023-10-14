@quarantined
Feature: Add student package
  Background:
    Given some package data in db

  Scenario: Unauthenticated user try to add student package
    Given an invalid authentication token
    When user add a "valid" package for a student
    Then returns "Unauthenticated" status code

  Scenario Outline: Authenticate user try to add student package with invalid case
    Given a valid authentication token
    When user add a "<packageId>" package for a student
    Then returns "<code>" status code

    Examples:
      | packageId        | code            |
      | empty            | InvalidArgument |
      | not exist        | NotFound        |
      | already inactive | InvalidArgument |

  Scenario: Authenticate user try to add student package with valid case
    Given a valid authentication token
    When user add a "already active" package for a student
    Then returns "OK" status code
    And server must store this package for this student

  Scenario: Authenticate user try to add course to student package with valid case
    Given a signed in "school admin"
    When user add a package by courses for a student
    Then returns "OK" status code
    And server must store these courses for this student

  Scenario: Authenticate user try to add course to student package with valid case
    Given a signed in "school admin"
    When user add a package by courses with student package extra for a student
    Then returns "OK" status code
    And server must store these courses and class for this student

