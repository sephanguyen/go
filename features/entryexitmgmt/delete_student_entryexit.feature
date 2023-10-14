@major
Feature: Delete Student Entry and Exit records
  As a school staff
  I am able to delete a student entry and exit record
  Background:
    Given there is an existing student

  Scenario Outline: School staff deletes entry and exit record successfully
    Given student has "<entry-exit>" record
    When "<signed-in user>" deletes that record of this student
    Then receives "OK" status code

    Examples:
      | signed-in user | entry-exit |
      | school admin   | entry      |
      | school admin   | exit       |
      | hq staff       | entry      |
      | hq staff       | exit       |
      | centre lead    | entry      |
      | centre lead    | exit       |
      | centre manager | entry      |
      | centre manager | exit       |
      | centre staff   | entry      |
      | centre staff   | exit       |
